package analyzer

import (
	"github.com/op/go-logging"
	mdw "chaoshen.com/sccrawler/middleware"
	"chaoshen.com/sccrawler/model"
	"errors"
	"fmt"
)

var logger= logging.MustGetLogger("analyzer")

var idGenerator mdw.IdGenerator=mdw.NewIdGenerator()

func genDownloaderID()uint32{
	return idGenerator.GetUint32()
}


type Analyzer interface {
	Id() uint32
	Analyzer(resp model.Response,parsers []ParseResponse)([]model.Request,[]model.ItemData,[]error)
}

type analyzerImp struct {
	id uint32
}

func (ali *analyzerImp)Id()uint32{
	return ali.id
}

func (ali *analyzerImp)Analyzer(resp model.Response,parsers []ParseResponse)([]model.Request,[]model.ItemData,[]error){
	if resp.Valid()==false{
		err:=errors.New("Response invalid!")
		return nil,nil,[]error{err}
	}
	if parsers==nil{
		err:=errors.New("Response parsers is nil!")
		return nil,nil,[]error{err}
	}
	logger.Debugf("Parse the response URL=%s,depth=%d",resp.HttpRep().Request.URL,resp.Depth())
	requestList:=make([]model.Request,0)
	itemList:=make([]model.ItemData,0)
	errorList:=make([]error,0)

	for i,parser:= range parsers {
		if parser==nil{
			errMsg:=fmt.Sprintf("The [%d]th parser is nil!",i)
			appendErrorList(errorList,[]error{errors.New(errMsg)})
		}
		reqRes,itemRes,errRes:=parser(resp)
		requestList=appendRequestList(requestList,reqRes)
		itemList=appendItemList(itemList,itemRes)
		errorList=appendErrorList(errorList,errRes)
	}
	return  requestList,itemList,errorList
}

func appendErrorList(errorList []error,errList2 []error) []error{
	for _,err:= range errList2{
		if err==nil{
			continue
		}
		errorList=append(errorList,err)
	}
	return errorList
}

func appendRequestList(resultList[]model.Request,partList []model.Request) []model.Request{
	for i,elem:=range partList{
		if !elem.Valid(){
			logger.Debugf("Invaild element %d",i)
			continue
		}
		resultList=append(resultList,elem)
	}
	return resultList
}


func appendItemList(resultList[]model.ItemData,partList []model.ItemData) []model.ItemData{
	for i,elem:=range partList{
		if !elem.Valid(){
			logger.Debugf("Invaild element %d",i)
			continue
		}
		resultList=append(resultList,elem)
	}
	return resultList
}

