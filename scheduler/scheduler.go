package scheduler

import (
	"net/http"
	"github.com/op/go-logging"
	"chaoshen.com/sccrawler/model"
	 anlz "chaoshen.com/sccrawler/analyzer"
	 ipl "chaoshen.com/sccrawler/itempipeline"
	mid "chaoshen.com/sccrawler/middleware"
	"chaoshen.com/sccrawler/downloader"
	"math"
	"fmt"
	"errors"
	"sync/atomic"
	"github.com/goc2p/src/webcrawler/base"
	"time"
	"go/token"
)

const (
	DOWNLOADER_CODE   = "downloader"
	ANALYZER_CODE     = "analyzer"
	ITEMPIPELINE_CODE = "item_pipeline"
	SCHEDULER_CODE    = "scheduler"
)

type GenHttpClient func() *http.Client

var logger=logging.MustGetLogger(SCHEDULER_CODE)


type Scheduler interface {
	//启动调度器
	Start(
		chanCfg model.ChannelConfig,
		pollCfg model.PoolBaseConfig,
		crawlDepth uint32,
		httpClientGenerator GenHttpClient,
		respParsers []anlz.ParseResponse,
		processors []ipl.ItemProcessor,
		firstRequest *http.Request,
	)(err error)
	//关闭调度器
	Stop() bool
	//是否在运行
	Running() bool
	//错误通道
	ErrorChan() <-chan error
	//Summary(prefix string) SchedSummary
}

func NewScheduler()Scheduler {
	return &SchedulerImp{}
}

type SchedulerImp struct{
	chanCfg model.ChannelConfig
	pollCfg model.PoolBaseConfig
	crawlDepth uint32
	primaryDomain string
	chanman mid.ChannelManager
	stopSign mid.StopSign
	dlPool  downloader.PageDownloaderPool
	alPool  anlz.AnalyzerPool
	itemPipe ipl.ItemPipeline
	reqCache requestCache
	urlMap  map[string]bool
	running uint32 // 运行标记。0表示未运行，1表示已运行，2表示已停止
}


func (sche *SchedulerImp)Start (
	chanCfg model.ChannelConfig,
	pollCfg model.PoolBaseConfig,
	crawlDepth uint32,
	httpClientGenerator GenHttpClient,
	respParsers []anlz.ParseResponse,
	processors []ipl.ItemProcessor,
	firstRequest *http.Request) (err error){

	defer func(){
		if p:=recover();p!=nil{
			errMsg:=fmt.Sprintf("Fatal Scheduler Error:%s\n",p)
			logger.Fatal(errMsg)
			err=errors.New(errMsg)
		}
	}()

	if atomic.LoadUint32(&sche.running)==1{
		return errors.New("The scheduler has been started!\n")
	}
	atomic.StoreUint32(&sche.running,1)
	if err:=chanCfg.Check();err!=nil{
		return err
	}
	sche.chanCfg=chanCfg

	if err:=pollCfg.Check();err!=nil{
		return err
	}
	sche.pollCfg=pollCfg
	sche.crawlDepth=crawlDepth

	if firstRequest == nil {
		return errors.New("The first HTTP request is invalid!")
	}

	if primaryDomain,err:=getPrimaryDomain(firstRequest.Host);err!=nil{
		return err
	}else{
		sche.primaryDomain=primaryDomain
	}


	sche.chanman=genChanmgr(&sche.chanCfg)

	if sche.stopSign==nil{
		sche.stopSign=mid.NewStopSign()
	}else{
		sche.stopSign.Reset()
	}

	if httpClientGenerator == nil {
		return errors.New("The HTTP client generator list is invalid!")
	}

	dlPool,err:=generatePageDownloaderPool(sche.pollCfg.PageDownloaderPoolSize(),httpClientGenerator)
	if err!=nil{
		errMsg:=fmt.Sprintf("Occur error when get page downloader pool %s\n!",err)
		errors.New(errMsg)

	}
	sche.dlPool=dlPool

	alPool,err:=generateAnalyzerPool(sche.pollCfg.AnalyzerPoolSize())
	if err!=nil{
		errMsg:=fmt.Sprintf("Occur error when get analyzer pool %s\n!",err)
		errors.New(errMsg)
	}
	sche.alPool=alPool


	if processors==nil{
		return errors.New("The item processor list is invalid!")
	}
	for i,ip:=range processors {
		if ip==nil{
			errMsg:=fmt.Sprintf("The item processor [%d] is invalid!",i)
			return errors.New(errMsg)
		}
	}
	sche.itemPipe=generateItemPipeline(processors)

	sche.reqCache=newRequestCache()

	sche.urlMap=make(map[string]bool)

	sche.startDownloading()
	sche.activateAnalyzers(respParsers)
	sche.openItemPipeline()
	sche.schedule(10 * time.Millisecond)

	firstReq:=model.NewRequest(firstRequest,0)
	sche.reqCache.put(firstReq)

	return nil
}


func (sche  *SchedulerImp)startDownloading(){
	go func(){
		for {
			req,ok:=<-sche.getReqChan()
			if !ok{
				break
			}
			go sche.download(req)
		}
	}()
}

func (sche  *SchedulerImp)download(request model.Request){
	defer func(){
		if p:=recover(); p!=nil{
			errMsg := fmt.Sprintf("Fatal Download Error: %s\n", p)
			logger.Fatal(errMsg)
		}
	}()
	downloader,err:=sche.dlPool.Take()
	if err!=nil{
		errMsg := fmt.Sprintf("Take downloader pool error: %s", err)
		sche.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func(){
		err:=sche.dlPool.Return(downloader)
		if err!=nil{
			errMsg := fmt.Sprintf("Return downloader pool error: %s", err)
			sche.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()

	response,err:=downloader.Download(request)
	if err!=nil{
		errMsg := fmt.Sprintf("Downloader pool error: %s", err)
		sche.sendError(errors.New(errMsg), SCHEDULER_CODE)
	}
	code:=generateCode(DOWNLOADER_CODE,downloader.Id())
	if response!=nil{
		sche.sendResp(response,code)
	}
	if err!=nil{
		sche.sendError(err,code)
	}
}

// 激活分析器。
func (sche *SchedulerImp) activateAnalyzers(respParsers []anlz.ParseResponse) {
	go func() {
		for {
			resp, ok := <-sche.getRespChan()
			if !ok {
				break
			}
			go sche.analyze(respParsers, resp)
		}
	}()
}


// 获取通道管理器持有的响应通道。
func (sche *SchedulerImp) getRespChan() chan model.Response {
	respChan, err := sche.chanman.GetRepChan()
	if err != nil {
		panic(err)
	}
	return respChan
}

func (sche *SchedulerImp) getReqChan() chan model.Request {
	respChan, err := sche.chanman.GetReqChan()
	if err != nil {
		panic(err)
	}
	return respChan
}