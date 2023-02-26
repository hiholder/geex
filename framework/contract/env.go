package contract

const (
	EnvProduction = "production"
	EnvTesting = "testing"
	EnvDevelopment = "development"
	EnvKey = "geex:env"
)

// Env 环境变量接口
type Env interface {
	// AppEnv 获取当前的环境
	AppEnv() string
	// IsExist 判断一个环境变量是否被配置
	IsExist(string) bool
	// Get 根据配置名获取相应的配置
	Get(string) string
	// All 获取全部配置，.env 和运行环境变量融合后结果
	All() map[string]string
}