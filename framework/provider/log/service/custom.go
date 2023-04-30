package service

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"io"
)

type GeexCustomLog struct {
	GeexLog
}

func NewGeexCustomLog(params ...interface{}) (interface{}, error) {
	c := params[0].(framework.Container)
	level := params[1].(contract.LogLevel)
	ctxFields := params[2].(contract.CtxFields)
	formatter := params[3].(contract.CtxFormatter)
	output := params[4].(io.Writer)
	log := &GeexCustomLog{}
	log.SetLevel(level)
	log.SetFields(ctxFields)
	log.SetFormatter(formatter)
	log.SetOutput(output)
	log.c = c
	return log, nil
}