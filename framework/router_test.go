package framework

import (
	c "github.com/smartystreets/goconvey/convey"
	"testing"
)

func init()  {
	e = New()
	e.addRouter("GET", "/", nil)
	e.addRouter("GET", "/hello/:name", nil)
	e.addRouter("GET", "/hello/b/c", nil)
	e.addRouter("GET", "/hi/:name", nil)
	e.addRouter("GET", "/assets/*filepath", nil)
}

func TestParsePattern(t *testing.T) {
	c.Convey("test router", t, func() {
		c.So(parsePattern("/p/:name"), c.ShouldResemble, []string{"p", ":name"})
		c.So(parsePattern("/p/*"), c.ShouldResemble, []string{"p", "*"})
		c.So(parsePattern("/p/*name/*"), c.ShouldResemble, []string{"p", "*name"})
	})
}

func TestGetRoute(t *testing.T) {
	c.Convey("Test Get Route", t, func() {
		handler, ps := e.methodTree["GET"].SearchRouter("/hello/geektutu")
		c.So(handler, c.ShouldNotBeNil)
		c.So(ps["name"], c.ShouldEqual, "geektutu")
	})
}
