package model

import (
	"testing"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("model/data")

func TestNewRequest(t *testing.T) {
	logger.Debugf("Into TestNewRequest")
	req:=NewRequest(nil,1)
	logger.Debugf("valid %t",req.Valid())
}