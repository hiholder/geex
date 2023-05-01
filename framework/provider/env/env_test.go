package env

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"github.com/hiholder/geex/framework/provider/app"
	c "github.com/smartystreets/goconvey/convey"
	"testing"
)
const (
	BasePath = ""
)
func TestGeexEnv_AppEnv(t *testing.T) {
	c.Convey("", t, func() {
		env, _ := NewGeexEnv("")
		geexEnv := env.(GeexEnv)
		c.So(geexEnv.maps["APP_ENV"], c.ShouldEqual, "development")
	})
}

func TestGeeXEnvProvider(t *testing.T) {
	c.Convey("test hade env normal case", t, func() {
		basePath := BasePath
		container := framework.NewGeeXContainer()
		sp := &app.GeexAppProvider{BaseFolder: basePath}

		err := container.Bind(sp)
		c.So(err, c.ShouldBeNil)

		sp2 := &GeexEnvProvider{}
		err = container.Bind(sp2)
		c.So(err, c.ShouldBeNil)

		envServ := container.MustMake(contract.EnvKey).(contract.Env)
		c.So(envServ.AppEnv(), c.ShouldEqual, "development")
		// So(envServ.Get("DB_HOST"), ShouldEqual, "127.0.0.1")
		// So(envServ.AppDebug(), ShouldBeTrue)
	})
}
