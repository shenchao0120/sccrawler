package middleware

import (
	"reflect"
	"sync"
	"fmt"
	"errors"
)

type PoolEntity interface {
	Id() uint32
}

type Pool interface {
	Take() (PoolEntity ,error)
	Return(poolEntity PoolEntity) error
	Total() uint32
	Used() uint32
}

func NewPool(total uint32, entityType reflect.Type, genEntity func() PoolEntity) (Pool,error){
	if total==0 {
		errMsg:=fmt.Sprintf("The pool can not be initialized! total=%d\n",total)
		return nil,errors.New(errMsg)
	}
	container:=make(chan PoolEntity,total)
	idContainer:=make(map[uint32]bool)
	for i:=0;i<int(total);i++{
		entity:=genEntity()
		if entityType!=reflect.TypeOf(entity){
			errMsg:=fmt.Sprintf("The type of result of function gentEntity is Not %s!\n",entityType)
			return nil,errors.New(errMsg)
		}
		container<-entity
		idContainer[entity.Id()]=true
	}
	pool:=&PoolImp{
		total:total,
		entityType:entityType,
		genEntity:genEntity,
		container:container,
		idContainer:idContainer,
	}
	return pool,nil
}

type PoolImp struct {
	total uint32 // capacity of pool
	entityType reflect.Type //type of  entity
	genEntity func() PoolEntity // gen entity function
	container chan PoolEntity  // channel of entity
	idContainer map[uint32]bool // ID indicator
	mutex sync.Mutex
}

func (pi *PoolImp)Take()(PoolEntity,error){
	entity,ok:=<-pi.container
	if !ok{
		errMsg:=fmt.Sprintf("The pool container is invalid!")
		return nil,errors.New(errMsg)
	}
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	pi.idContainer[entity.Id()]=false
	return entity,nil
}

func (pi *PoolImp)Return(entity PoolEntity) error{
	if entity==nil{
		return errors.New("The returning entity is invalid!")
	}
	if pi.entityType!=reflect.TypeOf(entity){
		errMsg:=fmt.Sprintf("The returning entity is not %s",pi.entityType)
		errors.New(errMsg)
	}
	entityId:=entity.Id()
	casResult:=pi.compareAndSetForIdContainer(entityId,false,true)
	if casResult==0 {
		pi.container<-entity
		return nil
	} else if casResult==1{
		errMsg:=fmt.Sprintf("The entity id=%d is already in the poll!\n",entityId)
		return errors.New(errMsg)
	} else {
		errMsg:=fmt.Sprintf("The entity id=%d is not illegal!\n",entityId)
		return errors.New(errMsg)
	}
}

func (pi *PoolImp)Total() uint32{
	return pi.total
}

func (pi *PoolImp)Used() uint32{
	return pi.total-uint32(len(pi.container))
}

func (pi *PoolImp)compareAndSetForIdContainer(entityId uint32,oldValue bool, newValue bool) int8{
	pi.mutex.Lock()
	defer pi.mutex.Unlock()
	v,ok:=pi.idContainer[entityId]
	if !ok{
		return -1
	}
	if v!=oldValue{
		return 1
	}
	pi.idContainer[entityId]=newValue
	return 0
}