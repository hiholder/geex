package framework

import (
	"net/http"
	"strings"
)

var (
	// http支持的方法
	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

type IGroup interface {
	Get(string, HandlerFunc) IGroup
	Post(string, HandlerFunc) IGroup
	Delete(string, HandlerFunc) IGroup
	Patch(string, HandlerFunc) IGroup
	Put(string, HandlerFunc) IGroup
	Options(string, HandlerFunc) IGroup
	Head(string, HandlerFunc) IGroup
	Group(string) IGroup
	Use(...HandlerFunc)
}

type Router struct {
	roots    map[string]*node
	Handlers map[string]HandlerFunc
}

// RouterGroup 路由组
type RouterGroup struct {
	prefix     string
	middleware []HandlerFunc // 中间件
	engine     *Engine       // 全局唯一的Engine
}

func newRouter() *Router {
	return &Router{
		roots:    make(map[string]*node),
		Handlers: make(map[string]HandlerFunc),
	}
}

func newGroup(e *Engine, prefix string) *RouterGroup {
	return &RouterGroup{
		engine: e,
		prefix: prefix,
	}
}

//func (r *Router) addRouter(method, pattern string, handler HandlerFunc) {
//	parts := parsePattern(pattern)
//	key := method + "-" + pattern
//	if _, ok := r.roots[method]; !ok {
//		r.roots[method] = &node{}
//	}
//	r.roots[method].insert(pattern, parts, 0)
//	r.Handlers[key] = handler
//}

//func (r *Router) getRouter(method, pattern string) (*node, map[string]string) {
//	searchParts := parsePattern(pattern)
//	params := make(map[string]string) // 用于解析通配符
//	if root, ok := r.roots[method]; ok {
//		if node := root.search(searchParts, 0); node != nil {
//			parts := parsePattern(node.pattern)
//			for index, part := range parts {
//				// 去除路由中的通配符，用来提取参数，比如":lang"输入路径为"username"时，需要在map中保存key为lang，value为username"
//				// 去除路由中的":"
//				if part[0] == ':' {
//					params[part[1:]] = searchParts[index]
//				}
//				// 去除路由中的"*"
//				if part[0] == '*' && len(part) > 1 {
//					params[part[1:]] = strings.Join(searchParts[index:], "/")
//					break
//				}
//			}
//			return node, params
//		}
//	}
//	return nil, nil
//}

//func (r *Router) FindRouteByRequest(c *Context) {
//	if node, params := r.getRouter(c.Method, c.Path); node != nil {
//		c.Params = params
//		key := c.Method + "-" + node.pattern
//		//r.Handlers[key](c)
//		c.handlers = append(c.handlers, r.Handlers[key]) // 将handler加入handler中
//	} else {
//		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
//	}
//	c.Next() // 运行
//}

func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}
