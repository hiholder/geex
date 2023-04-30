package contract

import (
	"context"
	"io"
	"time"
)

const LogKey = "geex:log"

type LogLevel int32

const (
	UnknownLevel LogLevel = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Log interface {
	CtxFatal(ctx context.Context, msg string, fields map[string]interface{})
	CtxError(ctx context.Context, msg string, fields map[string]interface{})
	CtxWarn(ctx context.Context, msg string, fields map[string]interface{})
	CtxInfo(ctx context.Context, msg string, fields map[string]interface{})
	CtxDebug(ctx context.Context, msg string, fields map[string]interface{})
	CtxTrace(ctx context.Context, msg string, fields map[string]interface{})
	SetLevel(level LogLevel)
	SetFields(fields CtxFields)
	SetFormatter(formatter CtxFormatter)
	// SetOutput 设置输出管道，具体的输出由配置决定
	SetOutput(writer io.Writer)
}
// CtxFields 获取日志上下文参数
type CtxFields func(ctx context.Context) map[string]interface{}

// CtxFormatter 控制日志的输出格式
type CtxFormatter func(level LogLevel, time time.Time, msg string, fields map[string]interface{}) ([]byte, error)


