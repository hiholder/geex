package framework

import (
	"net/http"
	"path"
)

// Static 解析请求地址，映射到服务器文件上的真实地址
func (r *RouterGroup) Static(relativePath, root string) {
	handler := r.createStatic(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	r.Get(urlPattern, handler)
}

func (r *RouterGroup) createStatic(relativePath string, fs http.FileSystem) HandlerFunc {
	// 构造绝对路径，例如: "v1/assets/index.js"
	absolutePath := path.Join(r.prefix, relativePath)
	// 创建文件服务器
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}
