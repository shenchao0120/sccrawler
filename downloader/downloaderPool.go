package downloader

import (
	mdw "chaoshen.com/sccrawler/middleware"
	"reflect"
	"fmt"
	"errors"
)
type GenPageDownloader func() PageDownloader
type PageDownloaderPool interface {
	Take()(PageDownloader,error)
	Return(pdl PageDownloader)error
	Total() uint32
	Used() uint32
}

func NewPageDownloaderPool(total uint32,gen GenPageDownloader)(PageDownloaderPool,error){
	entityType:=reflect.TypeOf(gen())
	genEntity:=func() mdw.PoolEntity{
		return gen()
	}
	pool,err:=mdw.NewPool(total,entityType,genEntity)
	if err!=nil{
		return nil,err
	}
	return &PageDownloaderPoolImp{pool:pool,entityType:entityType},nil
}

type PageDownloaderPoolImp struct {
	pool mdw.Pool
	entityType reflect.Type
}

func (pdpi *PageDownloaderPoolImp)Take() (PageDownloader,error){
	poolEntity,err:=pdpi.pool.Take()
	if err!=nil{
		return nil,err
	}
	pd,ok:=poolEntity.(PageDownloader)
	if !ok{
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", pdpi.entityType)
		panic(errors.New(errMsg))
	}
	return pd,nil
}


func (pdpi *PageDownloaderPoolImp)Return(pd PageDownloader)error{
	return pdpi.pool.Return(pd)
}

func (pdpi *PageDownloaderPoolImp)Total()uint32 {
	return pdpi.pool.Total()
}

func (pdpi *PageDownloaderPoolImp)Used()uint32 {
	return pdpi.pool.Used()
}


