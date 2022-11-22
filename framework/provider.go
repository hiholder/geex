package framework

type NewInstance func(...interface{}) (interface{}, error)


type ServiceProvider interface {
	Name()   string
	Register(Container) NewInstance
	Params(Container) []interface{}
	IsDefer() bool
	Boot(Container) error
}
