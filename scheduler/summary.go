package scheduler

import (
	"chaoshen.com/sccrawler/model"
	"fmt"
	"bytes"
)

type SchedSummary interface {
	String() string
	Detail() string
	Same(other SchedSummary)bool
}


type SchedSummaryImp struct {
	prefix string
	running uint32
	channelConfig model.ChannelConfig
	poolBaseConfig model.PoolBaseConfig
	crawlDepth uint32
	chanMgrSummary string
	reqCacheSummary string
	dlPoolLen uint32
	dlPoolCap uint32
	alPoolLen uint32
	alPoolCap uint32
	itemPipeLineSummary string
	urlCount int
	urlDetail    string
	stopSignSummary string
}

func NewSchedSummary(sched *SchedulerImp,prefix string)SchedSummary{
	if sched==nil {
		return nil
	}
	sched.urlMapMutex.RLock()
	urlCount:=len(sched.urlMap)
	var urlDetail string
	if urlCount>0{
		var buffer bytes.Buffer
		buffer.WriteByte('\n')
		for k,_:=range sched.urlMap {
			buffer.WriteString(prefix)
			buffer.WriteString(k)
			buffer.WriteByte('\n')
		}
		sched.urlMapMutex.RUnlock()

		urlDetail=buffer.String()
	}else{
		urlDetail="\n"
	}
	return &SchedSummaryImp{
		prefix:prefix,
		running:sched.running,
		channelConfig:sched.chanCfg,
		poolBaseConfig:sched.pollCfg,
		crawlDepth:sched.crawlDepth,
		chanMgrSummary:sched.chanman.Summary(),
		reqCacheSummary:sched.reqCache.summary(),
		dlPoolLen:sched.dlPool.Used(),
		dlPoolCap:sched.dlPool.Total(),
		alPoolLen:sched.alPool.Used(),
		alPoolCap:sched.alPool.Total(),
		itemPipeLineSummary:sched.itemPipe.Summary(),
		urlCount:urlCount,
		urlDetail:urlDetail,
		stopSignSummary:sched.stopSign.Summary(),
	}
}

func (ss *SchedSummaryImp) String() string {
	return ss.getSummary(false)
}

func (ss *SchedSummaryImp) Detail() string {
	return ss.getSummary(true)
}




// 获取摘要信息。
func (ss *SchedSummaryImp) getSummary(detail bool) string {
	prefix := ss.prefix
	template := prefix + "Running: %v \n" +
		prefix + "Channel config: %s \n" +
		prefix + "Pool base config: %s \n" +
		prefix + "Crawl depth: %d \n" +
		prefix + "Channels manager: %s \n" +
		prefix + "Request cache: %s\n" +
		prefix + "Downloader pool: %d/%d\n" +
		prefix + "Analyzer pool: %d/%d\n" +
		prefix + "Item pipeline: %s\n" +
		prefix + "Urls(%d): %s" +
		prefix + "Stop sign: %s\n"
	return fmt.Sprintf(template,
		func() bool {
			return ss.running == 1
		}(),
		ss.channelConfig.String(),
		ss.poolBaseConfig.String(),
		ss.crawlDepth,
		ss.chanMgrSummary,
		ss.reqCacheSummary,
		ss.dlPoolLen, ss.dlPoolCap,
		ss.alPoolLen, ss.alPoolCap,
		ss.itemPipeLineSummary,
		ss.urlCount,
		func() string {
			if detail {
				return ss.urlDetail
			} else {
				return "<concealed>\n"
			}
		}(),
		ss.stopSignSummary)
}


func (ss *SchedSummaryImp) Same(other SchedSummary) bool {
	if other == nil {
		return false
	}
	otherSs, ok := interface{}(other).(*SchedSummaryImp)
	if !ok {
		return false
	}
	if ss.running != otherSs.running ||
		ss.crawlDepth != otherSs.crawlDepth ||
		ss.dlPoolLen != otherSs.dlPoolLen ||
		ss.dlPoolCap != otherSs.dlPoolCap ||
		ss.alPoolCap != otherSs.alPoolCap ||
		ss.alPoolLen != otherSs.alPoolLen ||
		ss.urlCount != otherSs.urlCount ||
		ss.stopSignSummary != otherSs.stopSignSummary ||
		ss.reqCacheSummary != otherSs.reqCacheSummary ||
		ss.poolBaseConfig.String() != otherSs.poolBaseConfig.String() ||
		ss.channelConfig.String() != otherSs.channelConfig.String() ||
		ss.itemPipeLineSummary != otherSs.itemPipeLineSummary ||
		ss.chanMgrSummary != otherSs.chanMgrSummary {
		return false
	} else {
		return true
	}
}
