package model

import (
	"net/http"
	"net/url"
	"time"
	"chaoshen.com/sccrawler/otherTools"
)

//基础数据接口

type Data interface {
	Valid() bool //有效性
}


//http request 包装
type Request struct {
	httpReq *http.Request
	depth uint32
}

func NewRequest (httpReq *http.Request, depth uint32) *Request{
	return &Request{httpReq:httpReq,depth:depth}
}


func (req *Request)HttpReq() *http.Request{
	return req.httpReq
}

func (req *Request)Depth() uint32{
	return req.depth
}

func (req *Request)Valid() bool{
	return  req.httpReq!=nil && req.httpReq.URL!=nil
}

//http response 包装

type Response struct {
	httpRep *http.Response
	depth uint32
}

func NewResponse (httpRep *http.Response, depth uint32) *Response{
	return &Response{httpRep:httpRep,depth:depth}
}


func (rep *Response)HttpRep() *http.Response{
	return rep.httpRep
}

func (rep *Response)Depth() uint32{
	return rep.depth
}

func (rep *Response)Valid() bool{
	return  rep.httpRep!=nil && rep.httpRep.Body!=nil
}

// Item parsed from page

type Item struct {
	id  uint32
	url *url.URL
	depth uint32
	timestamp []byte
	dataMap ItemData
	rawData []byte
	ProcData []byte
	PipeErrs []error
}

func NewItem(req Response ) *Item {
	id :=genItemID()
	timestamp:=[]byte(time.Now().String())
	dataMap:=make(ItemData)
	return &Item{id:id,
		url:req.httpRep.Request.URL,
		depth:req.depth,
		timestamp:timestamp,
		dataMap:dataMap,
	}
}

func (item * Item)Valid() bool {
	return item.url!=nil && item.rawData!=nil
}

func (item * Item)URL() *url.URL{
	return item.url
}

func (item * Item)Depth() uint32{
	return item.depth
}

func (item * Item)Timestamp() []byte{
	return item.timestamp
}

func (item * Item)RawData() []byte{
	return item.rawData
}

func (item * Item)Id() uint32{
	return item.id
}

func (item * Item)AddDataMap(key string ,value interface{})bool{
	if key=="" || value ==nil{
		return false
	}
	if item.dataMap==nil{
		item.dataMap=make(ItemData)
	}
	item.dataMap[key]=value
	return true
}

func (item * Item)DataMap() ItemData{
	return item.dataMap
}

type ItemData map[string]interface{}


var idGenerator otherTools.IdGenerator=otherTools.NewIdGenerator()

func genItemID()uint32{
	return idGenerator.GetUint32()
}
