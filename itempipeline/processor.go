package itempipeline

import "chaoshen.com/sccrawler/model"

type ItemProcessor func(pItem *model.Item)(pResult *model.Item,err error)
