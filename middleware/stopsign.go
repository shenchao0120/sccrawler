package middleware

import (
	"sync"
	"fmt"
)

type StopSign interface {
	//sign the stop flag
	Sign() bool
	// is signed
	IsSigned() bool
	// reset
	Reset()
	//record dealt code
	Dealt(code string)
	// The number of dealt
	DealtCount(code string) uint32
	//Total number of dealt
	DealtTotal() uint32
	//Summary of stop sign
	Summary() string
}

func NewStopSign()StopSign{
	ss:=&StopSignImp{
		dealtCount:make(map[string]uint32),
	}
	return  ss
}

type StopSignImp struct {
	rwMutex sync.RWMutex
	stopSign bool   //true means stop, false mean not
	dealtCount map[string]uint32
}

func (ssi *StopSignImp)Sign() bool{
	ssi.rwMutex.Lock()
	defer ssi.rwMutex.Unlock()
	if ssi.stopSign==true{
		return false
	}
	ssi.stopSign=true
	return true
}

func (ssi *StopSignImp)IsSigned()bool{
	ssi.rwMutex.RLock()
	defer ssi.rwMutex.RUnlock()
	return ssi.stopSign

}

func (ssi *StopSignImp)Reset(){
	ssi.rwMutex.Lock()
	defer ssi.rwMutex.Unlock()
	ssi.stopSign=false
	return
}

func (ssi *StopSignImp)Dealt(code string){
	ssi.rwMutex.Lock()
	defer ssi.rwMutex.Unlock()
	if ssi.stopSign==false{
		return
	}
	if _,ok:=ssi.dealtCount[code];!ok {
		ssi.dealtCount[code]=1
	} else{
		ssi.dealtCount[code]++
	}
}

func (ssi *StopSignImp)DealtCount(code string)uint32{
	ssi.rwMutex.RLock()
	defer ssi.rwMutex.RUnlock()
	return ssi.dealtCount[code]
}

func (ssi *StopSignImp)DealtTotal()uint32{
	var total uint32
	ssi.rwMutex.RLock()
	defer ssi.rwMutex.RUnlock()
	for _,v:= range ssi.dealtCount{
		total+=v
	}
	return total
}

func (ssi *StopSignImp) Summary() string {
	if ssi.stopSign {
		return fmt.Sprintf("signed: true, dealCount: %v", ssi.dealtCount)
	} else {
		return "signed: false"
	}
}


