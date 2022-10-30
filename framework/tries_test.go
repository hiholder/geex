package framework

import (
	"fmt"
	"github.com/sanity-io/litter"
	c "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

func TestAddRouter(t *testing.T)  {
	c.Convey("test add router", t, func() {
		url := "/v1/student/add"
		tree := NewTree()
		err := tree.AddRouter(url, func(c *Context) {
			fmt.Println("hello world")
		})
		c.So(err, c.ShouldBeNil)
		c.So(nodeString(tree.root), c.ShouldEqual, "/v1/student/add")
	})
}

func TestSearchRouter(t *testing.T) {
	c.Convey("test search router", t, func() {
		url := "/v1/student/add"
		tree := NewTree()
		err := tree.AddRouter(url, func(c *Context) {
			fmt.Println("hello world")
		})
		c.So(err, c.ShouldBeNil)
		handler, _ := tree.SearchRouter(url)
		c.So(handler, c.ShouldNotBeNil)
		r, _ := http.NewRequest(http.MethodGet, "/v1", nil)
		handler(newContext(nil, r))
	})
}

func TestDynamicRouter(t *testing.T) {
	c.Convey("test dynamic router", t, func() {
		url1 := "/v1/student/:name/age/age"
		tree := NewTree()
		err := tree.AddRouter(url1, func(c *Context) {
			fmt.Println("student name")
		})
		c.So(err, c.ShouldBeNil)
		handler, params := tree.SearchRouter("/v1/student/tom/age/age")
		c.So(handler, c.ShouldNotBeNil)
		litter.Dump(params)
		r, _ := http.NewRequest(http.MethodGet, "/v1", nil)
		handler(newContext(nil, r))
	})
}

func TestNullRouter(t *testing.T)  {
	c.Convey("test null router", t, func() {
		url := "/"
		tree := NewTree()
		err := tree.AddRouter(url, func(c *Context) {
			fmt.Println("student name")
		})
		c.So(err, c.ShouldBeNil)
		//c.So(nodeString(tree.root), c.ShouldEqual, "/")
		handler, _ := tree.SearchRouter("/")
		c.So(handler, c.ShouldNotBeNil)
		r, _ := http.NewRequest(http.MethodGet, "/v1", nil)
		handler(newContext(nil, r))
	})
}

func nodeString(node *node) string {
	if node == nil {
		return ""
	}
	path := node.part
	if node.part != "" {
		path = "/" + node.part
	}
	for _, child := range node.children {
		if child != nil {
			path +=  nodeString(child)
			break
		}
	}
	return path
}
