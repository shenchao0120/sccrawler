package analyzer

import "chaoshen.com/sccrawler/model"

type ParseResponse func (response model.Response)([]model.Request,[]model.ItemData,[]error)