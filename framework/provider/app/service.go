package app

import (
	"flag"
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/util"
	gerrors "github.com/pkg/errors"
	"path/filepath"
)

type GeexApp struct {
	container framework.Container // 服务容器
	baseFolder string	// 基础路径
	appID      string   // 表示当前这个app的唯一id, 可以用于分布式锁等
	configMap map[string]string
}

func (g GeexApp) Version() string {
	return "0.0.1"
}

func (g GeexApp) BaseFolder() string {
	if g.baseFolder != "" {
		return g.baseFolder
	}
	var baseFolder string
	flag.StringVar(&baseFolder, "base_folder", "", "base_folder参数，默认为当前路径")
	flag.Parse()
	if g.baseFolder != "" {
		return g.baseFolder
	}
	return util.GetExecDirectory()
}

func (g GeexApp) ConfigFolder() string {
	return filepath.Join(g.BaseFolder(), "config")
}

func (g GeexApp) LogFolder() string {
	return filepath.Join(g.StorageFolder(), "log")
}

func (g GeexApp) ProviderFolder() string {
	return filepath.Join(g.BaseFolder(), "provider")
}

func (g GeexApp) MiddleFolder() string {
	return filepath.Join(g.BaseFolder(), "middle")
}

func (g GeexApp) CommandFolder() string {
	return filepath.Join(g.BaseFolder(), "command")
}

func (g GeexApp) RuntimeFolder() string {
	return filepath.Join(g.BaseFolder(), "runtime")
}

func (g GeexApp) TestFolder() string {
	return filepath.Join(g.BaseFolder(), "test")
}

func (g GeexApp) StorageFolder() string {
	return filepath.Join(g.BaseFolder(), "storage")
}

func NewGeexApp(params ...interface{}) (interface{}, error) {
	if len(params) != 2 {
		return nil, gerrors.New("params error")
	}
	container, ok := params[0].(framework.Container)
	if !ok {
		return nil, gerrors.Errorf("invalid container: %v", params[0])
	}
	baseFolder, ok := params[1].(string)
	if !ok {
		return nil, gerrors.Errorf("invalid baseFolder: %v", baseFolder)
	}
	return &GeexApp{baseFolder: baseFolder, container: container}, nil
}

// LoadAppConfig 加载App配置
func (g *GeexApp)LoadAppConfig(kv map[string]string)  {
	for k, v := range kv {
		g.configMap[k] = v
	}
}




