package log

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"github.com/hiholder/geex/framework/provider/log/formatter"
	"github.com/hiholder/geex/framework/provider/log/service"
	"io"
	"strings"
)

type GeexLogServiceProvider struct {
	framework.ServiceProvider
	Driver    string
	// 日志级别
	Level     contract.LogLevel
	// 日志输出格式方法
	CtxFields contract.CtxFields
	// 日志context上下文信息获取函数
	Formatter contract.CtxFormatter
	// 日志输出信息
	Output io.Writer
}

func (g *GeexLogServiceProvider) Name() string {
	return contract.LogKey
}

func (g *GeexLogServiceProvider) Register(c framework.Container) framework.NewInstance {
	if g.Driver == "" {
		config, err := c.Make(contract.ConfigKey)
		if err != nil {
			return service.NewGeexConsoleLog
		}
		cf := config.(contract.Config)
		g.Driver = strings.ToLower(cf.GetString("log.Driver"))
	}
	switch g.Driver {
	case "console":
		return service.NewGeexConsoleLog
	case "custom":
		return service.NewGeexCustomLog
	case "single":
		return service.NewGeexSingleLog
	case "rotate":
		return service.NewGeexRotateLog
	default:
		return service.NewGeexConsoleLog
	}
}

func (g *GeexLogServiceProvider) Params(c framework.Container) []interface{} {
	config := c.MustMake(contract.ConfigKey).(contract.Config)
	if g.Formatter == nil {
		g.Formatter = formatter.TextFormatter
		if config.IsExist("log.formatter") {
			format := config.GetString("log.formatter")
			if format == "text" {
				g.Formatter = formatter.TextFormatter
			} else if format == "json" {
				g.Formatter = formatter.JsonFormatter
			}
		}
	}
	if g.Level == contract.UnknownLevel {
		g.Level = contract.InfoLevel
		if config.IsExist("log.level") {
			g.Level = logLevel(config.GetString("log.level"))
		}
	}
	return []interface{}{c, g.Level, g.CtxFields, g.Formatter, g.Output}
}

func (g *GeexLogServiceProvider) IsDefer() bool {
	return false
}

func (g GeexLogServiceProvider) Boot(container framework.Container) error {
	return nil
}
var levelMap = map[string]contract.LogLevel {
	"fatal": contract.FatalLevel,
	"error": contract.ErrorLevel,
	"warn": contract.WarnLevel,
	"info": contract.InfoLevel,
	"debug": contract.DebugLevel,
	"trace": contract.TraceLevel,
}
func logLevel(config string) contract.LogLevel {
	config = strings.ToLower(config)
	level, ok := levelMap[config]
	if ok {
		return level
	}
	return contract.UnknownLevel
}



