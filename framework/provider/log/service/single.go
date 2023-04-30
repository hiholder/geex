package service

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"github.com/hiholder/geex/framework/util"
	gerrors "github.com/pkg/errors"
	"os"
	"path/filepath"
)

type GeexSingleLog struct {
	GeexLog
	// 日志存储目录
	folder	string
	// 日志文件名
	file 	string
	fd      *os.File
}

func NewGeexSingleLog(params ...interface{}) (interface{}, error) {
	c := params[0].(framework.Container)
	level := params[1].(contract.LogLevel)
	ctxFields := params[2].(contract.CtxFields)
	formatter := params[3].(contract.CtxFormatter)
	app := c.MustMake(contract.AppKey).(contract.App)
	config := c.MustMake(contract.ConfigKey).(contract.Config)
	log := &GeexSingleLog{}
	log.SetLevel(level)
	log.SetFields(ctxFields)
	log.SetFormatter(formatter)
	folder := app.LogFolder()
	if config.IsExist("log.folder") {
		folder = config.GetString("log.folder")
	}
	log.folder = folder
	if !util.Exists(folder) {
		os.MkdirAll(folder, os.ModePerm)
	}
	log.file = "geex.log"
	if config.IsExist("log.file") {
		log.file = config.GetString("log.file")
	}
	fd, err := os.OpenFile(filepath.Join(log.folder, log.file), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, gerrors.Wrap(err, "open log file err")
	}
	log.SetOutput(fd)
	log.c = c
	return log, nil
}


