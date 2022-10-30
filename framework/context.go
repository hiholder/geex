package framework

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type H map[string]interface{}

// Context 请求上下文
type Context struct {
	// origin objects
	Writer http.ResponseWriter
	Req    *http.Request
	// 请求信息
	Path   string
	Method string
	Params map[string]string // 动态路由参数
	// 结果信息
	StatusCode int
	// middleware
	handlers []HandlerFunc
	index    int // 记录进行到哪个中间件
	// engine指针
	engine *Engine
	// 写保护机制
	writerMux  *sync.Mutex
	// 超时标记位
	hasTimeout bool
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Path:   r.URL.Path,
		Method: r.Method,
		index:  -1,
		writerMux: &sync.Mutex{},
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) SetHasTimeout()  {
	c.hasTimeout = true
}

// 构造String响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 构造JSON响应
func (c *Context) JSON(code int, obj interface{}) {
	if c.hasTimeout {
		return
	}
	defer c.writerMux.Unlock()
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	c.writerMux.Lock()
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Data 构造Data响应
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 构造HTML响应
func (c *Context) HTML(code int, name string, data interface{}) {
	if c.hasTimeout {
		return
	}
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	// 拥有渲染模板的能力
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}

// Param 可以访问到解析的参数，比如可以通过c.Param("lang")方法获取到对应的值
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// Fail 请求失败
func (c *Context) Fail(code int, err string)  {
	c.index = len(c.handlers)
	c.JSON(code, H{
		"message" : err,
	})
}

func (c *Context) Done() <-chan struct{}{
	if c.Req == nil || c.Req.Context() == nil {
		return nil
	}
	return c.Req.Context().Done()
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	if c.Req == nil || c.Req.Context() == nil {
		return
	}
	return c.Req.Context().Deadline()
}
