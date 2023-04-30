package service

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"os"
)

type GeexConsoleLog struct {
	GeexLog
}

func NewGeexConsoleLog(params ...interface{}) (interface{}, error) {
	c := params[0].(framework.Container)
	level := params[1].(contract.LogLevel)
	ctxFields := params[2].(contract.CtxFields)
	ctxFormatter := params[3].(contract.CtxFormatter)
	log := &GeexConsoleLog{}
	log.SetFields(ctxFields)
	log.SetFormatter(ctxFormatter)
	log.SetLevel(level)
	log.SetOutput(os.Stdout)
	log.c = c
	return log, nil
}
