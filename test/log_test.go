package test

import (
	"context"
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"github.com/hiholder/geex/framework/provider/log"
	"github.com/hiholder/geex/framework/provider/log/formatter"
	"github.com/hiholder/geex/framework/provider/log/service"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestLogPrint(t *testing.T)  {
	convey.Convey("test info", t, func() {
		fields := make(map[string]interface{})
		var f  contract.CtxFields
		iLog, _ := service.NewGeexConsoleLog(framework.NewGeeXContainer(), contract.InfoLevel, f, contract.CtxFormatter(formatter.TextFormatter))
		infoLog := iLog.(contract.Log)
		fields["test"] = "test info log"
		infoLog.CtxInfo(context.Background(), "test log", fields)
	})
}

func TestLogWithContainer(t *testing.T)  {
	convey.Convey("test log with container", t, func() {
		c := framework.NewGeeXContainer()
		geexLog := &log.GeexLogServiceProvider{}
		instance := geexLog.Register(c)
		iLog, err := instance(geexLog.Params(c))
		convey.So(err, convey.ShouldBeNil)
		logger := iLog.(contract.Log)
		fields := make(map[string]interface{})
		fields["test"] = "test info log"
		logger.CtxInfo(context.Background(), "test log", fields)
	})
}
