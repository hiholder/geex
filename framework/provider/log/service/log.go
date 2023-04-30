package service

import (
	"context"
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"github.com/hiholder/geex/framework/provider/log/formatter"
	"io"
	"time"
)

// 用于实现日志的具体功能

type GeexLog struct {
	level      contract.LogLevel
	formatter  contract.CtxFormatter
	ctxFields  contract.CtxFields
	output     io.Writer
	c          framework.Container
}


func (log *GeexLog) SetLevel(level contract.LogLevel) {
	log.level = level
}

func (log *GeexLog) SetFields(fields contract.CtxFields) {
	log.ctxFields = fields
}

func (log *GeexLog) SetFormatter(formatter contract.CtxFormatter) {
	log.formatter = formatter
}

func (log *GeexLog) SetOutput(writer io.Writer) {
	log.output = writer
}

func (log *GeexLog) CtxError(ctx context.Context, msg string, fields map[string]interface{}) {
	log.logf(contract.ErrorLevel, ctx, msg, fields)
}

func (log *GeexLog) CtxInfo(ctx context.Context, msg string, fields map[string]interface{})  {
	log.logf(contract.InfoLevel, ctx, msg, fields)
}

func (log *GeexLog) CtxWarn(ctx context.Context, msg string, fields map[string]interface{})  {
	log.logf(contract.WarnLevel, ctx, msg, fields)
}

func (log *GeexLog) CtxFatal(ctx context.Context, msg string, fields map[string]interface{})  {
	log.logf(contract.FatalLevel, ctx, msg, fields)
}

func (log *GeexLog) CtxDebug(ctx context.Context, msg string, fields map[string]interface{})  {
	log.logf(contract.DebugLevel, ctx, msg, fields)
}

func (log *GeexLog) CtxTrace(ctx context.Context, msg string, fields map[string]interface{})  {
	log.logf(contract.TraceLevel, ctx, msg, fields)
}

func (log *GeexLog) logf(level contract.LogLevel, ctx context.Context, msg string, fields map[string]interface{}) error {
	// 判断日志级别
	if !log.IsLevelEnable(level) {
		return nil
	}
	// 将上下文参数填充到fields中
	if fd := log.ctxFields; fd != nil {
		t := log.ctxFields(ctx)
		for k, v := range t {
			fields[k] = v
		}
	}
	// 用log绑定的输出格式输出
	if log.formatter == nil {
		log.formatter = formatter.TextFormatter
	}
	logBy, err := log.formatter(level, time.Now(), msg, fields)
	if err != nil {
		return err
	}

	// 通过output输出
	log.output.Write(logBy)
	log.output.Write([]byte("\t\n"))
	return nil
}

func (log *GeexLog)IsLevelEnable(level contract.LogLevel) bool {
	return level <= log.level
}
