package framework

import (
	c "github.com/smartystreets/goconvey/convey"
	"testing"
)
var r = newRouter()
func init()  {
	r.addRouter("GET", "/", nil)
	r.addRouter("GET", "/hello/:name", nil)
	r.addRouter("GET", "/hello/b/c", nil)
	r.addRouter("GET", "/hi/:name", nil)
	r.addRouter("GET", "/assets/*filepath", nil)
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
		node, ps := r.getRouter("GET", "/hello/geektutu")
		c.So(node, c.ShouldNotBeNil)
		c.So(node.pattern, c.ShouldEqual, "/hello/:name")
		c.So(ps["name"], c.ShouldEqual, "geektutu")
	})
}
