package framework

type NewInstance func(...interface{}) (interface{}, error)


type ServiceProvider interface {
	// Name 代表了服务提供者的凭证
	Name()   string
	// Register 在服务容器中注册一个实例化方法
	Register(Container) NewInstance
	// Params 定义传递给NewInstance的参数
	Params(Container) []interface{}
	// IsDefer 是否延迟实例化
	IsDefer() bool
	// Boot 实例化前的准备工作，例如基础配置，初始化参数
	Boot(Container) error
}
