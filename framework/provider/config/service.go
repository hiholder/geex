package config

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/hiholder/geex/framework"
	"github.com/hiholder/geex/framework/contract"
	"github.com/mitchellh/mapstructure"
	gerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type GeexConfig struct {
	c        framework.Container // 容器
	folder   string	// 文件夹
	keyDelim string	// 路径分隔符
	lock     sync.RWMutex
	envMap   map[string]string	// 所有环境变量
	confMap  map[string]interface{}	// 配置文件结构
	confRaw  map[string][]byte	// 配置文件原始信息
}

// NewGeexConfig 初始化Config的方法
func NewGeexConfig(params ...interface{}) (interface{}, error) {
	// 参数提取
	container := params[0].(framework.Container)
	envFolder := params[1].(string)
	envMap    := params[2].(map[string]string)
	// 检查文件夹是否存在
	if _, err := os.Stat(envFolder); os.IsNotExist(err) {
		return nil, gerrors.New("folder " + envFolder + " not exist: " + err.Error())
	}
	// 实例化
	geexConfig := &GeexConfig{
		c : container,
		folder: envFolder,
		envMap: envMap,
		keyDelim: ".",
		confMap: make(map[string]interface{}),
		confRaw: make(map[string][]byte),
		lock: sync.RWMutex{},
	}
	// 读取每个文件
	dir, err := ioutil.ReadDir(envFolder)
	if err != nil {
		return nil, gerrors.WithStack(err)
	}
	for _, file := range dir {
		fileName := file.Name()
		err = geexConfig.loadConfigFile(envFolder, fileName)
		if err != nil {
			logrus.Errorf("load Config File err: %v", err)
			continue
		}
	}
	// 监控文件夹内的文件
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, gerrors.WithStack(err)
	}
	if err = watcher.Add(envFolder); err != nil {
		return nil, gerrors.WithStack(err)
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		for {
			select {
			case ev := <- watcher.Events:
				{
					path, _ := filepath.Abs(ev.Name)
					index := strings.LastIndex(path, string(os.PathSeparator))
					folder := path[:index]
					fileName := path[index+1:]
					if ev.Op&fsnotify.Create == fsnotify.Create {
						logrus.Infof("创建文件：%v", ev.Name)
						geexConfig.loadConfigFile(folder, fileName)
					}
					if ev.Op&fsnotify.Write == fsnotify.Write {
						logrus.Infof("写入文件： %v", ev.Name)
						geexConfig.loadConfigFile(folder, fileName)
					}
					if ev.Op&fsnotify.Rename == fsnotify.Rename {
						logrus.Infof("移除文件： %v", ev.Name)
						geexConfig.removeConfigFile(folder, fileName)
					}
				}
			case err := <- watcher.Errors:
				{
					logrus.Errorf("watch dir err: %v", err)
					return
				}
			}
		}
	}()
	return geexConfig, nil
}

func (conf *GeexConfig)loadConfigFile(folder string, file string) error {
	conf.lock.Lock()
	defer conf.lock.Unlock()
	s := strings.Split(file, ".")
	if len(s) != 2 || (s[1] == "yaml" || s[1] == "yml") {
		return nil
	}
	name := s[0]
	// 读取文件内容
	bf, err := ioutil.ReadFile(filepath.Join(folder, file))
	if err != nil {
		return gerrors.WithStack(err)
	}
	// 替换环境变量和配置文件
	bf = replace(bf, conf.envMap)
	// 解析对应文件
	var c map[string]interface{}
	if err = yaml.Unmarshal(bf, &c); err != nil {
		return gerrors.WithStack(err)
	}
	conf.confMap[name] = c
	conf.confRaw[name] = bf
	// 读取app.path的信息，更新app对应的folder
	if name == "app" && conf.c.IsBind(contract.AppKey) {
		if p, ok := c["path"]; ok {
			appService := conf.c.MustMake(contract.AppKey).(contract.App)
			appService.LoadAppConfig(cast.ToStringMapString(p))
		}
	}
	return nil
}

func (conf *GeexConfig)removeConfigFile(folder string, file string) error {
	conf.lock.Lock()
	defer conf.lock.Unlock()
	s := strings.Split(file, conf.keyDelim)
	if len(s) == 2 && (s[1] == "yaml" || s[1] == "yml") {
		fileName := s[0]
		// 删除对应的key
		delete(conf.confMap, fileName)
		delete(conf.confRaw, fileName)
	}
	return nil
}

func (conf *GeexConfig) IsExist(key string) bool {
	return conf.find(key) != nil
}

func (conf *GeexConfig) Get(key string) interface{} {
	return conf.find(key)
}

func (conf *GeexConfig) GetBool(key string) bool {
	return cast.ToBool(conf.find(key))
}

func (conf *GeexConfig) GetInt(key string) int {
	return cast.ToInt(conf.find(key))
}

func (conf *GeexConfig) GetFloat64(key string) float64 {
	return cast.ToFloat64(conf.find(key))
}

func (conf *GeexConfig) GetString(key string) string {
	return cast.ToString(conf.find(key))
}

func (conf *GeexConfig) GetTime(key string) time.Time {
	return cast.ToTime(conf.find(key))
}

func (conf *GeexConfig) GetIntSlice(key string) []int {
	return cast.ToIntSlice(conf.find(key))
}

func (conf *GeexConfig) GetStringSlice(key string) []string {
	return cast.ToStringSlice(conf.find(key))
}

func (conf *GeexConfig) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(conf.find(key))
}

func (conf *GeexConfig) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(conf.find(key))
}

func (conf *GeexConfig) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(conf.find(key))
}

func (conf *GeexConfig) Load(key string, val interface{}) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			TagName: "yaml",
			Result: val,
		})
	if err != nil {
		return gerrors.WithStack(err)
	}
	return decoder.Decode(conf.find(key))
}

func (conf *GeexConfig) find(key string) interface{} {
	conf.lock.RLock()
	defer conf.lock.RUnlock()
	return searchMap(conf.confMap, strings.Split(key, conf.keyDelim))
}

func replace(content []byte, env map[string]string) []byte {
	if len(env) == 0 {
		return content
	}
	for k, v := range env {
		key := "env(" + k + ")"
		content = bytes.ReplaceAll(content, []byte(key), []byte(v))
	}
	return content
}

func searchMap(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}
	next, ok := source[path[0]]
	if ok {
		switch next.(type) {
		case map[string]interface{}:
			return searchMap(next.(map[string]interface{}), path[1:])
		case map[interface{}]interface{}:
			return searchMap(cast.ToStringMap(next), path[1:])
		default:
			return nil
		}
	}
	return nil
}
