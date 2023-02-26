package framework

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/spf13/cast"
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
	writerMux *sync.Mutex
	// 超时标记位
	hasTimeout bool
	// 服务容器
	container  Container
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {

	return &Context{
		Writer:    w,
		Req:       r,
		Path:      r.URL.Path,
		Method:    r.Method,
		index:     -1,
		writerMux: &sync.Mutex{},
	}
}

// 获取表单数据

func (c *Context) FormAll() map[string][]string {
	if c.Req != nil {
		return c.Req.PostForm
	}
	return map[string][]string{}
}

func (c *Context) FormInt(key string, def int) (int, bool) {
	if v, ok := c.FormAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToInt(v[0]), true
		}
	}
	return def, false
}

func (c *Context) FormInt64(key string, def int64) (int64, bool) {
	if v, ok := c.FormAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToInt64(v[0]), true
		}
	}
	return def, false
}

func (c *Context) FormFloat32(key string, def float32) (float32, bool) {
	if v, ok := c.FormAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToFloat32(v[0]), true
		}
	}
	return def, false
}

func (c *Context) FormFloat64(key string, def float64) (float64, bool) {
	if v, ok := c.FormAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToFloat64(v[0]), true
		}
	}
	return def, false
}

func (c *Context) FormString(key string, def string) (string, bool) {
	if v, ok := c.FormAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToString(v[0]), true
		}
	}
	return def, false
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// 查询请求头数据

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) QueryAll() map[string][]string {
	return c.Req.URL.Query()
}

func (c *Context) QueryInt(key string, def int) (int, bool) {
	if v, ok := c.QueryAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToInt(v[0]), true
		}
	}
	return def, false
}

func (c *Context) QueryInt64(key string, def int64) (int64, bool) {
	if v, ok := c.QueryAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToInt64(v[0]), true
		}
	}
	return def, false
}

func (c *Context) QueryFloat32(key string, def float32) (float32, bool) {
	if v, ok := c.QueryAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToFloat32(v[0]), true
		}
	}
	return def, false
}

func (c *Context) QueryFloat64(key string, def float64) (float64, bool) {
	if v, ok := c.QueryAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToFloat64(v[0]), true
		}
	}
	return def, false
}

func (c *Context) QueryString(key string, def string) (string, bool) {
	if v, ok := c.QueryAll()[key]; ok {
		if len(v) > 0 {
			return cast.ToString(v[0]), true
		}
	}
	return def, false
}

// header
func (c *Context) Headers() map[string][]string {
	return c.Req.Header
}

func (c *Context) Header(key string) (string, bool) {
	vals := c.Req.Header.Values(key)
	if len(vals) <= 0 {
		return "", false
	}
	return vals[0], true
}

// cookie
func (c *Context) Cookies() map[string]string {
	cookies := c.Req.Cookies()
	res := make(map[string]string)
	for _, cookie := range cookies {
		res[cookie.Name] = cookie.Value
	}
	return res
}

func (c *Context) Cookie(key string) (string, bool) {
	cookies := c.Cookies()
	if v, ok := cookies[key]; ok {
		return v, true
	}
	return "", false
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) SetHasTimeout() {
	c.hasTimeout = true
}

// 构造相应
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
	// 拥有渲染模板的能力
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
	c.Status(code)
}

func (c *Context) Xml(code int, obj interface{}) {
	marshal, err := xml.Marshal(obj)
	if err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
	c.SetHeader("Content-Type", "application/html")
	_, err = c.Writer.Write(marshal)
	if err != nil {
		c.Fail(http.StatusInternalServerError, err.Error())
	}
	c.Status(code)
}

func (c *Context) Redirect(path string) {
	http.Redirect(c.Writer, c.Req, path, http.StatusMovedPermanently)
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
func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers)
	c.JSON(code, H{
		"message": err,
	})
}

func (c *Context) Done() <-chan struct{} {
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

// context实现container的封装
func (c *Context) Make(key string) (interface{}, error) {
	return c.container.Make(key)
}

func (c *Context)MustMake(key string) interface{}  {
	return c.container.MustMake(key)
}

func (c *Context)MakeNew(key string, params []interface{}) (interface{}, error)  {
	return c.container.MakeNew(key, params)
}


