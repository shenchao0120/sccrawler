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

func genAnalyzerId()uint32{
	return idGenerator.GetUint32()
}


type Analyzer interface {
	Id() uint32
	Analyze(resp model.Response,parsers []ParseResponse)([]model.Request,* model.Item,[]error)
}

type analyzerImp struct {
	id uint32
}


func NewAnalyzer() Analyzer {
	return &analyzerImp{id: genAnalyzerId()}
}

func (ali *analyzerImp)Id()uint32{
	return ali.id
}

func (ali *analyzerImp)Analyze(resp model.Response,parsers []ParseResponse)([]model.Request,* model.Item,[]error){
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
	errorList:=make([]error,0)
	item:=model.NewItem(resp)

	for i,parser:= range parsers {
		if parser==nil{
			errMsg:=fmt.Sprintf("The [%d]th parser is nil!",i)
			appendErrorList(errorList,[]error{errors.New(errMsg)})
		}
		reqRes,itemRes,errRes:=parser(resp)
		requestList=appendRequestList(requestList,reqRes)
		errorList=appendErrorList(errorList,errRes)
		for k,v:= range itemRes{
			if k!="" && v!=nil{
				item.AddDataMap(k,v)
			}
		}
	}
	return  requestList,item,errorList
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




