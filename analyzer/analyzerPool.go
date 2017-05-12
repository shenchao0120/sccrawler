package analyzer


import (
mdw "chaoshen.com/sccrawler/middleware"
"reflect"
"fmt"
"errors"
)

type GenAnalyzer func() Analyzer

type AnalyzerPool interface {
	Take()(Analyzer,error)
	Return(al Analyzer)error
	Total() uint32
	Used() uint32
}

func NewAnalyzerPool(total uint32,gen GenAnalyzer)(AnalyzerPool,error){
	entityType:=reflect.TypeOf(gen())
	genEntity:=func() mdw.PoolEntity{
		return gen()
	}
	pool,err:=mdw.NewPool(total,entityType,genEntity)
	if err!=nil{
		return nil,err
	}
	return &AnalyzerPoolImp{pool:pool,entityType:entityType},nil
}

type AnalyzerPoolImp struct {
	pool mdw.Pool
	entityType reflect.Type
}

func (alpi *AnalyzerPoolImp)Take() (Analyzer,error){
	poolEntity,err:=alpi.pool.Take()
	if err!=nil{
		return nil,err
	}
	al,ok:=poolEntity.(Analyzer)
	if !ok{
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", alpi.entityType)
		panic(errors.New(errMsg))
	}
	return al,nil
}


func (alpi *AnalyzerPoolImp)Return(al Analyzer)error{
	return alpi.pool.Return(al)
}

func (alpi *AnalyzerPoolImp)Total()uint32 {
	return alpi.pool.Total()
}

func (alpi *AnalyzerPoolImp)Used()uint32 {
	return alpi.pool.Used()
}


