package service

import (
	"fmt"
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"github.com/hiholder/geex/framework/util"
	"github.com/lestrrat-go/file-rotatelogs"
	gerrors "github.com/pkg/errors"
	"os"
	"path/filepath"
	"time"
)

type GeexRotateLog struct {
	GeexLog
	// 日志文件存储目录
	folder  string
	// 日志文件名
	file    string
}

func NewGeexRotateLog(params ...interface{}) (interface{}, error) {
	c := params[0].(framework.Container)
	level := params[1].(contract.LogLevel)
	ctxFields := params[2].(contract.CtxFields)
	formatter := params[3].(contract.CtxFormatter)
	log := &GeexRotateLog{}
	log.SetLevel(level)
	log.SetFields(ctxFields)
	log.SetFormatter(formatter)
	app := c.MustMake(contract.AppKey).(contract.App)
	config := c.MustMake(contract.ConfigKey).(contract.Config)
	folder := app.LogFolder()
	if config.IsExist("log.folder") {
		folder = config.GetString("log.folder")
	}
	if !util.Exists(folder) {
		os.Mkdir(folder, os.ModePerm)
	}
	log.folder = folder
	file := "geex.log"
	if config.IsExist("log.file") {
		file = config.GetString("log.file")
	}
	log.file = file
	// 从配置文件中获取date_format信息
	dateFormat := "%Y%m%d%H"
	if config.IsExist("date_format") {
		dateFormat = config.GetString("log.date_format")
	}
	// 从配置中获取其他相关信息
	linkName := rotatelogs.WithLinkName(filepath.Join(folder, file))
	options := []rotatelogs.Option{linkName}
	// rotate_size
	if config.IsExist("log.rotate_size") {
		rotateSize := config.GetInt("log.rotate_size")
		options = append(options, rotatelogs.WithRotationSize(int64(rotateSize)))
	}
	// rotate_count
	if config.IsExist("log.rotate_count") {
		rotateCount := config.GetInt("log.rotate_count")
		options = append(options, rotatelogs.WithRotationCount(uint(rotateCount)))
	}
	// max_age
	if config.IsExist("log.max_age") {
		if maxAgeParse, err := time.ParseDuration(config.GetString("log.max_age")); err != nil {
			options = append(options, rotatelogs.WithMaxAge(maxAgeParse))
		}
	}
	// rotate_time
	if config.IsExist("log.rotate_time") {
		if rotateTimeParse, err := time.ParseDuration(config.GetString("log.rotate_time")); err != nil {
			options = append(options, rotatelogs.WithRotationTime(rotateTimeParse))
		}
	}
	w, err := rotatelogs.New(fmt.Sprintf("%s.%s", filepath.Join(log.folder, log.file), dateFormat), options...)
	if err != nil {
		return nil, gerrors.Wrap(err, "new rotatelogs error")
	}
	log.SetOutput(w)
	log.c = c
	return log, nil
}