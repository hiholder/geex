package main
import "github.com/hiholder/geex/framework"

// 服务提供方示例

type ServiceProvideDemo struct {

}

func (sp *ServiceProvideDemo) Name() string {
	return demoKey
}

func (sp *ServiceProvideDemo)Register(container framework.Container) framework.NewInstance  {

	return sp.NewInstance
}

func (sp *ServiceProvideDemo)Params(container framework.Container) []interface{} {
	return []interface{}{container}
}

func (sp *ServiceProvideDemo) IsDefer() bool {
	return true
}

func (sp *ServiceProvideDemo) Boot(container framework.Container) error {
	return nil
}

func (sp *ServiceProvideDemo)NewInstance(params ...interface{}) (interface{}, error) {
	c, ok := params[0].(framework.Container)
	if !ok {
		return nil, nil
	}
	return &DemoService{
		c: c,
	}, nil
}

type DemoService struct {
	Service
	c framework.Container
}

func (s *DemoService)GetFoo() Foo {
	return Foo{
		Name: "geex demo",
	}
}
