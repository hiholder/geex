package contract

import "time"

const (
	// ConfigKey is config key in container
	ConfigKey = "geex:config"
)
// Config 定义了配置文件服务，读取配置文件，支持点分割的路径读取
type Config interface {
	// IsExist 配置是否存在
	IsExist(key string) bool
	// Get 根据key获取对应属性值
	Get(key string) interface{}
	GetBool(key string) bool
	GetInt(key string) int
	GetFloat64(key string) float64
	GetString(key string) string
	GetTime(key string) time.Time
	GetIntSlice(key string) []int
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringMapStringSlice(key string) map[string][]string
	AddRemoteProvider(provider, endpoint, path string) error
	GetRemoteConfig() error
	// Load 加载到某个对象
	Load(key string, val interface{}) error
}
