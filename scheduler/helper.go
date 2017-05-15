package scheduler

import (
	"chaoshen.com/sccrawler/model"
	"chaoshen.com/sccrawler/middleware"
	"chaoshen.com/sccrawler/downloader"
	"chaoshen.com/sccrawler/analyzer"
	ipl "chaoshen.com/sccrawler/itempipeline"
	"fmt"
	"strings"
	"regexp"
	"errors"
)

func genChanmgr ( chanCfg *model.ChannelConfig)(middleware.ChannelManager){
	return middleware.NewChannelManager(chanCfg)
}

func generatePageDownloaderPool(poolSize uint32,gen GenHttpClient)(downloader.PageDownloaderPool,error){
	dlpool,err:=downloader.NewPageDownloaderPool(poolSize,func ()downloader.PageDownloader{
		return downloader.NewPageDownloader(gen())
	})
	if err!=nil{
		return nil,err
	}else {
		return dlpool,nil
	}
}


func generateAnalyzerPool(poolSize uint32)(analyzer.AnalyzerPool,error){
	alyz,err:=analyzer.NewAnalyzerPool(poolSize,func()analyzer.Analyzer{
		return analyzer.NewAnalyzer()
	})
	if err!=nil{
		return nil,err
	} else{
		return alyz,nil
	}
}

func generateItemPipeline(itemProcessors []ipl.ItemProcessor) ipl.ItemPipeline {
	return ipl.NewItemPipeline(itemProcessors)
}


// 生成组件实例代号。
func generateCode(prefix string, id uint32) string {
	return fmt.Sprintf("%s-%d", prefix, id)
}

// 解析组件实例代号。
func parseCode(code string) []string {
	result := make([]string, 2)
	var codePrefix string
	var id string
	index := strings.Index(code, "-")
	if index > 0 {
		codePrefix = code[:index]
		id = code[index+1:]
	} else {
		codePrefix = code
	}
	result[0] = codePrefix
	result[1] = id
	return result
}

var regexpForIp = regexp.MustCompile(`((?:(?:25[0-5]|2[0-4]\d|[01]?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|[01]?\d?\d))`)

var regexpForDomains = []*regexp.Regexp{
	// *.xx or *.xxx.xx
	regexp.MustCompile(`\.(com|com\.\w{2})$`),
	regexp.MustCompile(`\.(gov|gov\.\w{2})$`),
	regexp.MustCompile(`\.(net|net\.\w{2})$`),
	regexp.MustCompile(`\.(org|org\.\w{2})$`),
	// *.xx
	regexp.MustCompile(`\.me$`),
	regexp.MustCompile(`\.biz$`),
	regexp.MustCompile(`\.info$`),
	regexp.MustCompile(`\.name$`),
	regexp.MustCompile(`\.mobi$`),
	regexp.MustCompile(`\.so$`),
	regexp.MustCompile(`\.asia$`),
	regexp.MustCompile(`\.tel$`),
	regexp.MustCompile(`\.tv$`),
	regexp.MustCompile(`\.cc$`),
	regexp.MustCompile(`\.co$`),
	regexp.MustCompile(`\.\w{2}$`),
}

func getPrimaryDomain(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("The host is empty!")
	}
	if regexpForIp.MatchString(host) {
		return host, nil
	}
	var suffixIndex int
	for _, re := range regexpForDomains {
		pos := re.FindStringIndex(host)
		if pos != nil {
			suffixIndex = pos[0]
			break
		}
	}
	if suffixIndex > 0 {
		var pdIndex int
		firstPart := host[:suffixIndex]
		index := strings.LastIndex(firstPart, ".")
		if index < 0 {
			pdIndex = 0
		} else {
			pdIndex = index + 1
		}
		return host[pdIndex:], nil
	} else {
		return "", errors.New("Unrecognized host!")
	}
}
