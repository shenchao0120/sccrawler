package crawler

import (
	"github.com/op/go-logging"
	"chaoshen.com/sccrawler/model"
	"errors"
	"time"
	"fmt"
	"net/url"
	"io"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"net/http"
	"chaoshen.com/sccrawler/analyzer"
	"chaoshen.com/sccrawler/itempipeline"
	"chaoshen.com/sccrawler/scheduler"
)

var logger=logging.MustGetLogger("Demo Main")

func ItemProcessor(pItem *model.Item)(*model.Item, error){
	if pItem == nil {
		return nil, errors.New("Invalid item!")
	}
	time.Sleep(10 * time.Millisecond)
	return pItem, nil
}

func parseForATag(response model.Response)([]model.Request,model.ItemData,[]error){
	if response.HttpRep().Status!="200"{
		err := errors.New(
			fmt.Sprintf("Unsupported status code %d. (httpResponse=%v)", response.HttpRep()))
		return nil, nil,[]error{err}
	}
	var reqUrl *url.URL=response.HttpRep().Request.URL
	var httpRespBody io.ReadCloser=response.HttpRep().Body
	defer func() {
		if httpRespBody!=nil{
			httpRespBody.Close()
		}
	}()
	newRequest:=make([]model.Request,0)
	itemData:=make(map[string]interface{})
	errs:=make([]error,0)

	doc,err:=goquery.NewDocumentFromReader(httpRespBody)
	if err!=nil{
		return nil,nil,[]error{err}
	}
	// 查找“A”标签并提取链接地址
	doc.Find("a").Each(func(index int,sel *goquery.Selection){
		href,exists:=sel.Attr("href")
		if !exists || href =="" || href =="#" || href=="/"{
			return
		}
		href=strings.TrimSpace(href)
		lowerHref:=strings.ToLower(href)
		if href!="" && !strings.HasPrefix(lowerHref,"javascript"){
			aUrl,err:=url.Parse(href)
			if err!=nil{
				errs=append(errs,err)
				return
			}
			if !aUrl.IsAbs(){
				aUrl=reqUrl.ResolveReference(aUrl)
			}
			httpReq,err:=http.NewRequest("GET",aUrl.String(),nil)
			if err!=nil{
				errs=append(errs,err)
			}else{
				newReq:=model.Request{httpReq,response.Depth()}
				newRequest=append(newRequest,newReq)
			}
		}
		text:=strings.TrimSpace(sel.Text())
		if text!=""{
			imap := make(map[string]interface{})
			imap["parent_url"]=reqUrl
			imap["a.txt"]=text
			imap["a.index"]=index
			itemData[href]=&imap
		}
	})
	return newRequest,itemData,errs
}

func getResponseParsers()[]analyzer.ParseResponse{
	parsers:=[]analyzer.ParseResponse{
		parseForATag,
	}
	return parsers
}

func getItemProcessors() []itempipeline.ItemProcessor {
	itemProcessors := []itempipeline.ItemProcessor{
		ItemProcessor,
	}
	return itemProcessors
}

func genHttpClient() *http.Client {
	return &http.Client{}
}

func main(){
	sched:=scheduler.NewScheduler()
	//interval:=10*time.Millisecond
	chancfg:=model.NewChannelConfig(10,10,10,10)
	poolCfg:=model.NewPoolBaseConfig(3,3)
	crawlDepth := uint32(1)
	httpClientGenerator := genHttpClient
	respParsers := getResponseParsers()
	itemProcessors := getItemProcessors()
	startUrl := "http://www.sogou.com"
	firstHttpReq, err := http.NewRequest("GET", startUrl, nil)
	if err != nil {
		logger.Error(err)
		return
	}
	sched.Start(*chancfg,*poolCfg,crawlDepth,httpClientGenerator,respParsers,itemProcessors,firstHttpReq)



}

