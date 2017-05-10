package model

import (
	"errors"
	"fmt"
)

type Config interface {
	Check() error //参数检测
	String() string //字符串表示
}


type ChannelConfig struct {
	reqChanLen uint32
	repChanLen uint32
	itemsChanLen uint32
	errorChanLen uint32
	description string
}

var channelConfigTemplate string="{ reqChanLen: %d, respChanLen: %d," +
	" itemsChanLen: %d, errorChanLen: %d }"

func NewChannelConfig(reqChanLen uint32, repChanLen uint32, itemsChanLen uint32, errorChanLen uint32) *ChannelConfig{
	return &ChannelConfig{
		reqChanLen:reqChanLen,
		repChanLen:repChanLen,
		itemsChanLen:itemsChanLen,
		errorChanLen:errorChanLen}
}

func (chanCfg *ChannelConfig)ReqChanLen() uint32{
	return chanCfg.reqChanLen
}

func (chanCfg *ChannelConfig)RepChanLen() uint32{
	return chanCfg.repChanLen
}

func (chanCfg *ChannelConfig)ItemsChanLen() uint32{
	return chanCfg.itemsChanLen
}

func (chanCfg *ChannelConfig)ErrorChanLen() uint32{
	return chanCfg.errorChanLen
}

func (chanCfg *ChannelConfig) Check() error{
	if chanCfg.reqChanLen==0{
		return errors.New("The request channel max length (capacity) can not be 0!\n")
	}
	if chanCfg.repChanLen==0{
		return errors.New("The response channel max length (capacity) can not be 0!\n")
	}
	if chanCfg.itemsChanLen==0{
		return errors.New("The items channel max length (capacity) can not be 0!\n")
	}
	if chanCfg.errorChanLen==0 {
		return errors.New("The error channel max length (capacity) can not be 0!\n")
	}
	return nil
}

func (chanCfg *ChannelConfig) String() string{
	if chanCfg.description==""{
		chanCfg.description=fmt.Sprintf(
			channelConfigTemplate,
			chanCfg.reqChanLen,
			chanCfg.repChanLen,
			chanCfg.itemsChanLen,
			chanCfg.errorChanLen)
	}
	return chanCfg.description
}



type PoolBaseConfig struct {
	pageDownloaderPoolSize uint32
	analyzerPoolSize uint32
	description string
}


func NewPoolBaseConfig(pageDownloaderPoolSize uint32,analyzerPoolSize uint32) *PoolBaseConfig{
	return &PoolBaseConfig{pageDownloaderPoolSize:pageDownloaderPoolSize,analyzerPoolSize:analyzerPoolSize}
}

func (pbc *PoolBaseConfig) PageDownloaderPoolSize() uint32{
	return pbc.pageDownloaderPoolSize
}

func (pbc *PoolBaseConfig) AnalyzerPoolSize() uint32{
	return pbc.analyzerPoolSize
}


func (pbc *PoolBaseConfig)Check() error{
	if pbc.pageDownloaderPoolSize==0{
		errors.New("The page downloader pool size can not be 0!\n")
	}
	if pbc.analyzerPoolSize==0{
		errors.New("The analyzer pool size can not be 0!\n")
	}
	return nil
}


var poolBaseConfigTemplate string="{ pageDownloaderPoolSize: %d, analyzerPoolSize: %d}"

func (pbc *PoolBaseConfig)String() string{
	if pbc.description==""{
		pbc.description=fmt.Sprintf(poolBaseConfigTemplate,pbc.pageDownloaderPoolSize,pbc.analyzerPoolSize)
	}
	return pbc.description
}







