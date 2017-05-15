package itempipeline

import (
	"chaoshen.com/sccrawler/model"
	"sync/atomic"
	"errors"
	"fmt"
)

type ItemPipeline interface {
	// send the item to the process pipe
	Send(item *model.Item) []error
	//fastfail fail the process immediate
	FailFast() bool
	//set failFast
	SetFailFast (failFast bool)
	//counts:sended ,recevied ,processed
	Count() []uint64
	// count of processing items
	ProcessingNum() uint64
	//summary
	Summary() string
}


type ItemPipelineImp struct {
	processors []ItemProcessor
	failFast bool
	sent uint64
	accept uint64
	processed uint64
	processing uint64
}

func (ipl *ItemPipelineImp) Send(item *model.Item) []error{
	atomic.AddUint64(&ipl.processing,1)
	defer atomic.AddUint64(&ipl.processing,^uint64(0))
	atomic.AddUint64(&ipl.sent,1)
	errs:=make([]error,0)

	if item==nil{
		errs=append(errs,errors.New("The input item is nil!"))
		return errs
	}
	atomic.AddUint64(&ipl.accept,1)
	var err error
	for i,processor:= range ipl.processors{
		if item==nil{
			errs=append(errs,errors.New("item is nil after processing"+string(i)))
			break
		}
		item,err=processor(item)
		if err!=nil && ipl.failFast==true{
			errs=append(errs,err)
			break
		}
	}
	atomic.AddUint64(&ipl.processed,1)
	return errs
}

func (ipl *ItemPipelineImp) FailFast() bool{
	return ipl.failFast
}

func (ipl *ItemPipelineImp) SetFailFast (failFast bool) {
	ipl.failFast=failFast
}

func (ipl *ItemPipelineImp) Count() []uint64{
	counts:=make([]uint64,3)
	counts[0]=atomic.LoadUint64(&ipl.sent)
	counts[1]=atomic.LoadUint64(&ipl.accept)
	counts[2]=atomic.LoadUint64(&ipl.processed)
	return counts
}

func  (ipl *ItemPipelineImp) ProcessingNum() uint64 {
	return atomic.LoadUint64(&ipl.processing)
}

var summaryTemplate = "failFast: %v, processorNumber: %d," +
	" sent: %d, accepted: %d, processed: %d, processingNumber: %d"

func (ipl *ItemPipelineImp) Summary() string {
	counts := ipl.Count()
	summary := fmt.Sprintf(summaryTemplate,
		ipl.failFast, len(ipl.processors),
		counts[0], counts[1], counts[2], ipl.ProcessingNum())
	return summary
}

// 创建条目处理管道。
func NewItemPipeline(itemProcessors []ItemProcessor) ItemPipeline {
	if itemProcessors == nil {
		panic(errors.New(fmt.Sprintln("Invalid item processor list!")))
	}
	innerItemProcessors := make([]ItemProcessor, 0)
	for i, ip := range itemProcessors {
		if ip == nil {
			panic(errors.New(fmt.Sprintf("Invalid item processor[%d]!\n", i)))
		}
		innerItemProcessors = append(innerItemProcessors, ip)
	}
	return &ItemPipelineImp{processors: innerItemProcessors}
}