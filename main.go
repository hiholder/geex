package main

import (
	"fmt"
	"github.com/hiholder/geex/framework"
	"html/template"
	"log"
	"net/http"
	"time"
)

func onlyForV2() framework.HandlerFunc {
	return func(c *framework.Context) {
		t := time.Now()
		c.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	e := framework.Default()
	e.SetFuncMap(template.FuncMap{
		"FormatAsDate" : FormatAsDate,
	})
	e.LoadHTMLGlob("templates/*")
	e.Static("/assets", "./static")
	e.Get("/hello", func(c *framework.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})
	e.Get("/hello/:name", func(c *framework.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})
	e.Get("/assets/*filepath", func(c *framework.Context) {
		c.JSON(http.StatusOK, framework.H{
			"filepath" : c.Param("filepath"),
		})
	})
	e.Post("/login", func(c *framework.Context) {
		c.JSON(http.StatusOK, framework.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
	v1 := e.Group("/v1")
	{
		v1.Get("/hello", func(c *framework.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := e.Group("/v2")
	v2.Use(onlyForV2())
	{
		v2.Get("/hello/:name", func(c *framework.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.Post("/login", func(c *framework.Context) {
			c.JSON(http.StatusOK, framework.H{
				"username" : c.PostForm("username"),
				"password" : c.PostForm("password"),
			})
		})
	}
	e.Get("/", func(c *framework.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	e.Get("/date", func(c *framework.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", framework.H{
			"title": "geex",
			"now":   time.Now(),
		})
	})
	e.Get("/panic", func(c *framework.Context) {
		names := []string{"geek"}
		c.String(http.StatusOK, names[100])
	})
	e.Run(":9999")
}
