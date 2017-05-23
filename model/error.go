package model

import (
	"bytes"
	"fmt"
)

//自定义错误类型
type ErrorType string

const (
	DOWNLOADER_ERROR ErrorType="Downloader Error"
	ANALYZER_ERROR ErrorType="Analyzer Error"
	ITEMS_PROCESSOR_ERROR ErrorType="Items Processor Error"
	MIDDLEWARE_ERROR ErrorType="Middleware Error"
	SCHEDULER_ERROR ErrorType="Scheduler Error"
)


type CrawlerError interface {
	Type() ErrorType
	Error() string
}

type crawlerErrorImp struct {
	errType ErrorType
	errMsg string
	fullErrMsg string
}

func NewCrawlerError(errType ErrorType,errMsg string) CrawlerError{
	return &crawlerErrorImp{errType:errType, errMsg:errMsg}
}
func (cei *crawlerErrorImp)Type() ErrorType{
	return cei.errType
}

func (cei *crawlerErrorImp)Error()string{
	if cei.fullErrMsg==""{
		cei.getFullErrMsg()
	}
	return cei.fullErrMsg
}

func (cei *crawlerErrorImp)getFullErrMsg(){
	var buffer bytes.Buffer
	buffer.WriteString("Crawler Error:")
	if cei.errType!=""{
		buffer.WriteString(string(cei.errType))
		buffer.WriteByte(':')
	}
	buffer.WriteString(cei.errMsg)
	cei.fullErrMsg=fmt.Sprintf("%s\n",buffer.String())
	return
}