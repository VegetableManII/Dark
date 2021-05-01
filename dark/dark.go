package dark

import (
	"net/http"
)
// HandleFunc 定义请求处理句柄
type HandleFunc func(c *Context)

// Engine 继承自Handler接口实现ServeHTTP方法
type Engine struct {
	router *router
}
// New 引擎的构造函数
func New() *Engine {
	return &Engine{router:newRouter()}
}

func (e *Engine) addRoute(method string,pattern string,handler HandleFunc) {
	e.router.addRoute(method,pattern,handler)
}
// Get 处理HTTP GET请求
func (e *Engine) Get(pattern string,handler HandleFunc) {
	e.addRoute("GET",pattern,handler)
}
// POST 处理HTTP POST请求
func (e *Engine) POST(pattern string,handler HandleFunc) {
	e.addRoute( "POST",pattern,handler)
}
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr,e)
}
func (e *Engine) ServeHTTP(w http.ResponseWriter,r *http.Request) {
	c := newContext(w,r)
	e.router.handle(c)
}