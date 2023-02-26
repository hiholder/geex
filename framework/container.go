package framework

import (
	gerrors "github.com/pkg/errors"
	"sync"
)

// Container 服务容器，实现绑定服务获取服务
type Container interface {
	// Bind 绑定服务提供者
	Bind(provider ServiceProvider) error
	// IsBind 根据关键字判断是否绑定服务
	IsBind(key string) bool
	// Make 根据关键字获取服务
	Make(key string) (interface{}, error)
	// MustMake 根据关键字获取服务，如果关键字为绑定无法，会panic
	// 使用这个方法的时候要保证该服务已被绑定
	MustMake(key string) interface{}
	// MakeNew 根据参数数组创建服务
	MakeNew(key string, params []interface{}) (interface{}, error)
}

type GeeXContainer struct {
	Container
	providerMap map[string]ServiceProvider
	instanceMap map[string]interface{}
	mu          sync.RWMutex
}

func NewGeeXContainer() *GeeXContainer {
	return &GeeXContainer{
		providerMap: make(map[string]ServiceProvider),
		instanceMap: make(map[string]interface{}),
		mu: sync.RWMutex{},
	}
}

func (gxc *GeeXContainer) Bind(provider ServiceProvider) error {
	gxc.mu.Lock()
	defer gxc.mu.Unlock()
	gxc.providerMap[provider.Name()] = provider
	if !provider.IsDefer() {
		if err := provider.Boot(gxc); err != nil {
			return gerrors.Wrap(err, "Boot failed")
		}
		params := provider.Params(gxc)
		methods := provider.Register(gxc)
		instance, err := methods(params)
		if err != nil {
			return gerrors.Wrap(err, "register failed")
		}
		gxc.instanceMap[provider.Name()] = instance
	}
	return nil
}

func (gxc *GeeXContainer)IsBind(key string) bool {
	_, ok := gxc.providerMap[key]
	return ok
}

func (gxc *GeeXContainer) Make(key string) (interface{}, error) {
	return gxc.make(key, false, nil)
}


func (gxc *GeeXContainer) MustMake(key string) (instance  interface{}) {
	instance, _ = gxc.make(key, false, nil)
	return instance
}

func (gxc *GeeXContainer)MakeNew(key string, params []interface{}) (interface{}, error) {
	return gxc.make(key, true, params)
}

func (gxc *GeeXContainer) make(key string, force bool, params []interface{}) (interface{}, error) {
	gxc.mu.Lock()
	defer gxc.mu.Unlock()
	sp, ok := gxc.providerMap[key]
	if !ok {
		return nil, gerrors.New("")
	}
	if force {
		return gxc.newInstance(sp, nil)
	}
	if instance, ok := gxc.instanceMap[key]; ok {
		return instance, nil
	}
	instance, err := gxc.newInstance(sp, nil)
	if err != nil {
		return nil, err
	}
	gxc.instanceMap[key] = instance
	return instance, nil
}

func (gxc *GeeXContainer) newInstance(sp ServiceProvider, params []interface{}) (interface{}, error) {
	if err := sp.Boot(gxc);  err != nil {
		return nil, gerrors.Wrapf(err, "boot failed")
	}
	if params == nil {
		params = sp.Params(gxc)
	}
	init := sp.Register(gxc)
	inst, err := init(params...)
	if err != nil {
		return nil, gerrors.Wrapf(err, "new instance failed")
	}
	return inst, nil
}
