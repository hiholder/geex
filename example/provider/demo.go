package main

const demoKey = "geex:demo"

type Service interface {
	GetFoo() Foo
}

// Foo 服务接口定义的一个数据结构
type Foo struct {
	Name string
}
