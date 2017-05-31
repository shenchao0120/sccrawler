package downloader

import (
	"github.com/op/go-logging"
	"chaoshen.com/sccrawler/model"
	"net/http"
	"errors"
	"chaoshen.com/sccrawler/otherTools"
)

var logger= logging.MustGetLogger("Downloader")


var idGenerator otherTools.IdGenerator=otherTools.NewIdGenerator()

func genDownloaderID()uint32{
	return idGenerator.GetUint32()
}

type PageDownloader interface {
	Id() uint32
	Download(request model.Request)(* model.Response,error)
}

func NewPageDownloader(client *http.Client) PageDownloader {
	id := genDownloaderID()
	if client == nil {
		client = &http.Client{}
	}
	return &pageDownloaderImp{
		id:         id,
		httpClient: *client,
	}
}


type pageDownloaderImp struct {
	id uint32
	httpClient http.Client
}

func (pdi *pageDownloaderImp)Id() uint32{
	return pdi.id
}

func (pdi *pageDownloaderImp)Download(request model.Request)(*model.Response,error){
	if request.Valid()==false {
		return nil,errors.New("Invaild Request!")
	}
	logger.Debugf("Begin http request with url:%s\n",request.HttpReq().URL.String())
	if httpRep,err:=pdi.httpClient.Do(request.HttpReq());err!=nil{
		return nil,err
	} else {
		return model.NewResponse(httpRep,request.Depth()),nil
	}
}


