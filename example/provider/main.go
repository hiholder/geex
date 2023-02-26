package main

import (
	"github.com/hiholder/geex/framework"
	"log"
	"net/http"
)

func main()  {
	engine := framework.New()
	engine.Bind(&ServiceProvideDemo{})
	engine.Get("/subject/list/all", SubjectListController)
	engine.Run(":8888")
}

func SubjectListController(c *framework.Context)  {
	demoService, ok := c.MustMake(demoKey).(Service)
	if !ok {
		log.Printf("not implement")
		panic(c.MustMake(demoKey))
	}
	foo := demoService.GetFoo()
	c.JSON(http.StatusOK, framework.H{
		"demo": foo,
	})
}
