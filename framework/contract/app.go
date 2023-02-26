package contract

const AppKey = "gexx:app"

// App 定义接口, App需要实现该接口
type App interface {
	// Version 定义当前版本
	Version() string
	// BaseFolder 定义项目的基础地址
	BaseFolder() string
	// ConfigFolder 定义配置文件路径
	ConfigFolder() string
	// LogFolder 定义了日志所在的路径
	LogFolder() string
	// ProviderFolder 定义业务自己的服务提供者地址
	ProviderFolder() string
	// MiddleFolder 中间件
	MiddleFolder() string
	// CommandFolder 定义业务定义的命令
	CommandFolder()  string
	// RuntimeFolder 定义业务的运行
	RuntimeFolder()  string
	// TestFolder 存放测试所需的信息
	TestFolder()  string
	// LoadAppConfig 加载App配置
	LoadAppConfig(map[string]string)
}
