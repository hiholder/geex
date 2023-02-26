package config

import (
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"path/filepath"
)

type GeexConfigProvider struct {
	
}

func (g GeexConfigProvider) Name() string {
	return contract.ConfigKey
}

func (g GeexConfigProvider) Register(container framework.Container) framework.NewInstance {
	return NewGeexConfig
}

func (g *GeexConfigProvider) Params(container framework.Container) []interface{} {
	appService := container.MustMake(contract.AppKey).(contract.App)
	envService := container.MustMake(contract.EnvKey).(contract.Env)
	env := envService.AppEnv()
	// 配置文件夹地址
	configFolder := appService.ConfigFolder()
	envFolder := filepath.Join(configFolder, env)
	return []interface{}{container, envFolder, envService.All()}
}

func (g GeexConfigProvider) IsDefer() bool {
	return false
}

func (g GeexConfigProvider) Boot(container framework.Container) error {
	return nil
}
 