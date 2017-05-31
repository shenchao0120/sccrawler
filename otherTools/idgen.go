package otherTools

import (
	"sync"
	"math"
)

type IdGenerator interface {
	GetUint32() uint32
}

func NewIdGenerator() IdGenerator{
	return &cyclicIdGenerator{}
}

type cyclicIdGenerator struct {
	sn uint32
	ended bool
	mutex sync.Mutex
}

func (cycIdGen *cyclicIdGenerator)GetUint32() uint32{
	cycIdGen.mutex.Lock()
	defer cycIdGen.mutex.Unlock()
	if cycIdGen.ended{
		cycIdGen.sn=1
		cycIdGen.ended=false
		return cycIdGen.sn
	}
	id:=cycIdGen.sn
	if id<math.MaxUint32{
		cycIdGen.sn++
	} else{
		cycIdGen.ended=true
	}
	return id
}