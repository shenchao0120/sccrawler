package model

import (
	"net/http"
	"net/url"
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

//

type ItemData struct {
	id  uint32
	url *url.URL
	depth uint32
	timestamp []byte
	rawData []byte
	ProcData []byte
	PipeErrs []error
}

func (itemData * ItemData)Valid() bool {
	return itemData.url!=nil && itemData.rawData!=nil
}

func (itemData * ItemData)URL() *url.URL{
	return itemData.url
}

func (itemData * ItemData)Depth() uint32{
	return itemData.depth
}

func (itemData * ItemData)Timestamp() []byte{
	return itemData.timestamp
}

func (itemData * ItemData)RawData() []byte{
	return itemData.rawData
}

func (itemData * ItemData)ID() uint32{
	return itemData.id
}



type Items map[string]*ItemData


func (items Items)Valid() bool{
	return items!=nil
}









