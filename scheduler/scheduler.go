package scheduler

import (
	"net/http"
	"github.com/op/go-logging"
	"chaoshen.com/sccrawler/model"
	 anlz "chaoshen.com/sccrawler/analyzer"
	 ipl "chaoshen.com/sccrawler/itempipeline"
	mid "chaoshen.com/sccrawler/middleware"
	"chaoshen.com/sccrawler/downloader"
	"fmt"
	"errors"
	"sync/atomic"
	"time"
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

func (sche *SchedulerImp)Stop()bool{
	if atomic.LoadUint32(&sche.running)!=1{
		return false
	}
	sche.chanman.Close()
	sche.stopSign.Sign()
	sche.reqCache.close()

	atomic.StoreUint32(&sche.running,1)
	return true
}

func (sche* SchedulerImp) Running()bool{
	return atomic.LoadUint32(&sche.running)==1
}

func (sche* SchedulerImp)ErrorChan() <-chan error {
	if sche.chanman.Status()!=mid.CHANNEL_MANAGER_STATUS_INITIALIZED{
		return nil
	}
	if errChan,err:=sche.chanman.GetErrorChan();err!=nil{
		panic(err)
	}else{
		return errChan
	}
}

func (sche *SchedulerImp) Idle() bool {
	idleDlPool:=sche.dlPool.Used()==0
	idleAlPool:=sche.alPool.Used()==0
	idleItemPipeLine:=sche.itemPipe.ProcessingNum()==0

	return idleDlPool && idleAlPool && idleItemPipeLine
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

// 发送错误。
func (sche *SchedulerImp) sendError(err error, code string) bool {
	if err == nil {
		return false
	}
	codePrefix := parseCode(code)[0]
	var errorType model.ErrorType
	switch codePrefix {
	case DOWNLOADER_CODE:
		errorType = model.DOWNLOADER_ERROR
	case ANALYZER_CODE:
		errorType = model.ANALYZER_ERROR
	case ITEMPIPELINE_CODE:
		errorType = model.ITEMS_PROCESSOR_ERROR
	}
	cError := model.NewCrawlerError(errorType, err.Error())
	if sche.stopSign.IsSigned() {
		sche.stopSign.Dealt(code)
		return false
	}
	go func() {
		sche.ErrorChan() <- cError
	}()
	return true
}

// 分析。
func (sche *SchedulerImp) analyze(respParsers []anlz.ParseResponse, resp model.Response) {
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal Analysis Error: %s\n", p)
			logger.Fatal(errMsg)
		}
	}()
	analyzer, err := sche.alPool.Take()
	if err != nil {
		errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
		sche.sendError(errors.New(errMsg), SCHEDULER_CODE)
		return
	}
	defer func() {
		err := sche.alPool.Return(analyzer)
		if err != nil {
			errMsg := fmt.Sprintf("Analyzer pool error: %s", err)
			sche.sendError(errors.New(errMsg), SCHEDULER_CODE)
		}
	}()
	code := generateCode(ANALYZER_CODE, analyzer.Id())
	reqList,item, errs := analyzer.Analyze(resp,respParsers)
	if reqList!=nil{
		for _,req:=range reqList{
			if req.HttpReq()!=nil{
				sche.saveReqToCache(req,code)
			}
		}
	}

	if item!=nil{
		sche.sendItem(item,code)
	}

	if errs != nil {
		for _, err := range errs {
			sche.sendError(err, code)
		}
	}
}

func (sche *SchedulerImp) saveReqToCache() {
