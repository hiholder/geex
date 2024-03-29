package framework

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type HandlerFunc func(*Context)

type Engine struct {
	*RouterGroup
	groups     []*RouterGroup   // 存储所有的路由组
	methodTree map[string]*Tree // 为每个方法构建一棵路由树
	// 模板渲染
	htmlTemplates *template.Template // 模板
	funcMap       template.FuncMap   // 自定义模板渲染函数
	container     Container
}

func New() *Engine {
	engine := &Engine{}
	engine.RouterGroup = newGroup(engine, "")
	engine.methodTree = make(map[string]*Tree, 0)
	engine.groups = []*RouterGroup{engine.RouterGroup}
	engine.container = NewGeeXContainer()
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (e *Engine) addRouter(method, pattern string, handler HandlerFunc) {
	if _, ok := e.methodTree[method]; !ok {
		e.methodTree[method] = NewTree()
	}
	if err := e.methodTree[method].AddRouter(pattern, handler); err != nil {
		fmt.Println(err)
	}
}

func (r *RouterGroup) Get(pattern string, handler HandlerFunc) IGroup {
	return r.addRouter(http.MethodGet, pattern, handler)
}

func (r *RouterGroup) Post(pattern string, handler HandlerFunc) IGroup {
	return r.addRouter(http.MethodPost, pattern, handler)
}

func (r *RouterGroup) Delete(pattern string, handler HandlerFunc) IGroup {
	return r.addRouter(http.MethodDelete, pattern, handler)
}

func (r *RouterGroup) Patch(pattern string, handler HandlerFunc) IGroup {
	return r.addRouter(http.MethodPatch, pattern, handler)
}
func (r *RouterGroup) Put(pattern string, handler HandlerFunc) IGroup {
	return r.addRouter(http.MethodPut, pattern, handler)
}

func (r *RouterGroup) Options(pattern string, handler HandlerFunc) IGroup {
	return r.addRouter(http.MethodOptions, pattern, handler)
}

func (r *RouterGroup) Head(pattern string, handler HandlerFunc) IGroup {
	return r.addRouter(http.MethodHead, pattern, handler)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) newContext(w http.ResponseWriter, r *http.Request) *Context {
	c := newContext(w, r)
	c.container = e.container
	return c
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		// 判断请求符合哪些中间件
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middleware...)
		}
	}
	fmt.Println(r.URL)
	c := e.newContext(w, r)
	c.handlers = middlewares
	c.engine = e
	e.handleServeHTTP(c)
}

// Bind 绑定服务容器
func (e *Engine) Bind(provider ServiceProvider) error {
	return e.container.Bind(provider)
}

func (e *Engine) IsBind(key string) bool {
	return e.container.IsBind(key)
}

func (e *Engine) handleServeHTTP(ctx *Context) {
	handler, params := e.methodTree[ctx.Method].SearchRouter(ctx.Path)
	if handler != nil {
		ctx.Params = params
		handler(ctx)
	} else {
		fmt.Println("not match router")
	}
}

// Group 创建一个新分组并注册入Engine
func (r *RouterGroup) Group(prefix string) IGroup {
	nGroup := newGroup(r.engine, prefix)
	r.engine.groups = append(r.engine.groups, nGroup)
	return nGroup
}

// 为分组添加路由
func (r *RouterGroup) addRouter(method, comp string, handler HandlerFunc) IGroup {
	pattern := r.prefix + comp
	r.engine.addRouter(method, pattern, handler)
	return r
}

func (r *RouterGroup) Use(middlewares ...HandlerFunc) {
	r.middleware = append(r.middleware, middlewares...)
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}
