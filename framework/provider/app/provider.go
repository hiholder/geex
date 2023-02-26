package app

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
)

type GeexAppProvider struct {
	BaseFolder  string
}

func (g *GeexAppProvider) Name() string {
	return contract.AppKey
}

func (g *GeexAppProvider) Register(container framework.Container) framework.NewInstance {
	return NewGeexApp
}

func (g *GeexAppProvider) Params(container framework.Container) []interface{} {
	return []interface{}{container, g.BaseFolder}
}

func (g *GeexAppProvider) IsDefer() bool {
	return false
}

func (g *GeexAppProvider) Boot(container framework.Container) error {
	return nil
}



