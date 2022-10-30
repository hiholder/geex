package framework

import (
	"log"
	"net/http"
	"testing"
	"time"
)

var e *Engine

func init() {
	e = New()
}

func onlyForV2() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		c.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func TestSimpleGee(t *testing.T) {
	e.Use(Logger())
	e.Get("/", func(c *Context) {
		c.HTML(http.StatusOK, "", "<h1>Hello Gee</h1>")
	})
	e.Get("/hello", func(c *Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})
	e.Get("/hello/:name", func(c *Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})
	e.Get("/assets/*filepath", func(c *Context) {
		c.JSON(http.StatusOK, H{
			"filepath" : c.Param("filepath"),
		})
	})
	e.Post("/login", func(c *Context) {
		c.JSON(http.StatusOK, H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
	v1 := e.Group("/v1")
	{
		v1.Get("/", func(c *Context) {
			c.HTML(http.StatusOK, "", "<h1>Hello Gee</h1>")
		})
		v1.Get("/hello", func(c *Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := e.Group("v2")
	v2.Use(onlyForV2())
	{
		v2.Get("/hello/:name", func(c *Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.Post("/login", func(c *Context) {
			c.JSON(http.StatusOK, H{
				"username" : c.PostForm("username"),
				"password" : c.PostForm("password"),
			})
		})
	}
	e.Run(":9999")
}
