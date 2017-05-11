package middleware

import (
	. "chaoshen.com/sccrawler/model"
	"sync"
	"fmt"
	"errors"
)

type channelManagerStatus uint8


const (
	CHANNEL_MANAGER_STATUS_UNINITIALIZED channelManagerStatus=0 // 未初始化
	CHANNEL_MANAGER_STATUS_INITIALIZED   channelManagerStatus=1 // 初始化
	CHANNEL_MANAGER_STATUS_CLOSED	     channelManagerStatus=2 // 已关闭
)

var channelManagerStatusName =map[channelManagerStatus]string{
	CHANNEL_MANAGER_STATUS_UNINITIALIZED:"uninitialized",
	CHANNEL_MANAGER_STATUS_INITIALIZED:"initialized",
	CHANNEL_MANAGER_STATUS_CLOSED:"closed",
}

type ChannelManager interface{
	//初始化通道管理器
	Init(channelConfig *ChannelConfig,reset bool) bool
	//关闭
	Close()bool
	//获取请求通道
	GetReqChan()(chan Request,error)
	//获取响应通道
	GetRepChan()(chan Response,error)
	//获取数据处理通道
	GetItemsChan()(chan Items,error)
	//获取错误通道
	GetErrorChan()(chan error,error)
	//获取通道管理器状态
	Status()channelManagerStatus
	//获取通道管理器摘要
	Summary()string
}

func NewChannelManager(channelConfig *ChannelConfig) ChannelManager {
	chanmgr:=&channelManagerImp{}
	chanmgr.Init(channelConfig,true)
	return chanmgr
}

type channelManagerImp struct{
	channelConfig	*ChannelConfig
	reqChan		chan Request
	repChan		chan Response
	itemsChan	chan Items
	errChan		chan error
	status 		channelManagerStatus
	rwMutex 	sync.RWMutex
}


func (chanmgr *channelManagerImp)Init(channelConfig *ChannelConfig,reset bool) bool{
	if err:=channelConfig.Check();err!=nil{
		panic(err)
	}
	chanmgr.rwMutex.Lock()
	defer chanmgr.rwMutex.Unlock()
	if chanmgr.status==CHANNEL_MANAGER_STATUS_INITIALIZED && reset!=true {
		return false
	}
	chanmgr.status=CHANNEL_MANAGER_STATUS_INITIALIZED
	chanmgr.channelConfig=channelConfig
	chanmgr.reqChan=make(chan Request,channelConfig.ReqChanLen())
	chanmgr.repChan=make(chan Response,channelConfig.RepChanLen())
	chanmgr.itemsChan=make(chan Items,channelConfig.ItemsChanLen())
	chanmgr.errChan=make(chan error,channelConfig.ErrorChanLen())
	return true
}

func (chanmgr *channelManagerImp) Close() bool {
	chanmgr.rwMutex.Lock()
	defer chanmgr.rwMutex.Unlock()
	if chanmgr.status!=CHANNEL_MANAGER_STATUS_INITIALIZED{
		return false
	}
	close(chanmgr.reqChan)
	close(chanmgr.repChan)
	close(chanmgr.itemsChan)
	close(chanmgr.errChan)
	chanmgr.status=CHANNEL_MANAGER_STATUS_CLOSED
	return true
}
//checkStatus 并不是并发安全的，所以调用方需要保证并发安全
func (chanmgr *channelManagerImp)checkStatus() error{
	if chanmgr.status==CHANNEL_MANAGER_STATUS_INITIALIZED{
		return nil
	}
	statusName,ok:=channelManagerStatusName[chanmgr.status]
	if !ok{
		statusName=fmt.Sprintf("%d",chanmgr.status)
	}
	errMsg:=fmt.Sprintf("The Undesirable status of channel manager:%s!\n",statusName)
	return errors.New(errMsg)
}

func (chanmgr *channelManagerImp)GetReqChan()(chan Request,error){
	chanmgr.rwMutex.RLock()
	defer  chanmgr.rwMutex.RUnlock()
	if err:=chanmgr.checkStatus();err!=nil{
		return nil,err
	}
	return chanmgr.reqChan,nil
}

func (chanmgr *channelManagerImp)GetRepChan()(chan Response,error){
	chanmgr.rwMutex.RLock()
	defer  chanmgr.rwMutex.RUnlock()
	if err:=chanmgr.checkStatus();err!=nil{
		return nil,err
	}
	return chanmgr.repChan,nil
}

func (chanmgr *channelManagerImp)GetItemsChan()(chan Items,error){
	chanmgr.rwMutex.RLock()
	defer  chanmgr.rwMutex.RUnlock()
	if err:=chanmgr.checkStatus();err!=nil{
		return nil,err
	}
	return chanmgr.itemsChan,nil
}

func (chanmgr *channelManagerImp)GetErrorChan()(chan error,error){
	chanmgr.rwMutex.RLock()
	defer  chanmgr.rwMutex.RUnlock()
	if err:=chanmgr.checkStatus();err!=nil{
		return nil,err
	}
	return chanmgr.errChan,nil
}

func (chanmgr *channelManagerImp)Status() channelManagerStatus{
	return chanmgr.status
}

var chanmgrSummaryTemplate="status:%s,"+
	"requestChannel:%d/%d,"+
	"responseChannel:%d/%d,"+
	"itensChannel:%d/%d,"+
	"errorChannel:%d/%d"


func (chanmgr *channelManagerImp)Summary()string{
	return fmt.Sprintf(chanmgrSummaryTemplate,
		channelManagerStatusName[chanmgr.status],
		len(chanmgr.reqChan),cap(chanmgr.reqChan),
		len(chanmgr.repChan),cap(chanmgr.repChan),
		len(chanmgr.itemsChan),cap(chanmgr.itemsChan),
		len(chanmgr.errChan),cap(chanmgr.errChan),
	)
}