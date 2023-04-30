# geex
一个go语言编写的web框架，基于极客兔兔的《Go语言动手写Web框架》和极客时间的《手把手带你写一个Web框架》。

## 启动服务器
```go
// 启动框架默认服务器
e := framework.Default()
```
## 路由结构
内容详解：https://ecj6xn2q3r.feishu.cn/docx/WhdWdAkiXoO0lQxVt3ycCVUkn9b
使用了前缀树作为路由解析的结构
* 添加路由
```go
e := framework.Default()
e.Get("/assets/*filepath", func(c *framework.Context) {
//...
})

```
* 路由组
```go
v1 := e.Group("/v1")
	{
		v1.Get("/hello", func(c *framework.Context) {
			//...
		})
	}
```
* 路由参数解析<br>
参数解析功能实现了`:param`和`*`参数的解析
```go
e.Get("/hello/:name", func(c *framework.Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
e.Get("/hello/student/*", func(c *framework.Context) {
           //...
})
```
## WebSocket功能支持
该功能的实现基本移植了nhooyr/websocket的功能，正在完成中

## 实现服务容器
* 框架提供服务容器功能
* 每个服务都是一个服务提供者
* 每个服务都需要想框架注册
* 由框架控制服务的实例化
* 调用服务功能时从容器获取服务示例，再调用服务的具体功能