package env

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
)

type GeexEnvProvider struct {
	Folder string
}

func (g *GeexEnvProvider) Name() string {
	return contract.EnvKey
}

func (g *GeexEnvProvider) Register(container framework.Container) framework.NewInstance {
	return NewGeexEnv
}

func (g *GeexEnvProvider) Params(container framework.Container) []interface{} {
	return []interface{}{g.Folder}
}

func (g *GeexEnvProvider) IsDefer() bool {
	return false
}

func (g *GeexEnvProvider) Boot(container framework.Container) error {
	appService := container.MustMake(contract.AppKey).(contract.App)
	g.Folder = appService.BaseFolder()
	return nil
}

