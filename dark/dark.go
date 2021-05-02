package dark

import (
	"net/http"
	"strings"
)

// 路由分组
type RouterGroup struct {
	prefix      string
	middlewares []HandleFunc //支持中间件
	engine      *Engine      // 所有分组共享一个engin实例
}

// HandleFunc 定义请求处理句柄
type HandleFunc func(c *Context)

// Engine 继承自Handler接口实现ServeHTTP方法
type Engine struct {
	*RouterGroup
	// 继承RouterGroup可以像使用RouterGroup一样使用Engine
	router *router
	groups []*RouterGroup
}

// New 引擎的构造函数
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	// 添加第一个分组
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (e *Engine) addRoute(method string, pattern string, handler HandleFunc) {
	e.router.addRoute(method, pattern, handler)
}

// Get 处理HTTP GET请求
func (e *Engine) Get(pattern string, handler HandleFunc) {
	e.addRoute("GET", pattern, handler)
}

// POST 处理HTTP POST请求
func (e *Engine) POST(pattern string, handler HandleFunc) {
	e.addRoute("POST", pattern, handler)
}
func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// ServeHTTP接口吗，所有的HTTP请求都会通过该函数进入处理
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandleFunc
	for _, group := range e.groups {
		// 通过URL前缀判断属于哪一个路由分组的中间件
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, r)
	c.handlers = middlewares
	e.router.handle(c)
}

// 路由分组
// Group 创建一个路由分组对象，所有分组对象共享一个engine
func (g *RouterGroup) Group(prefix string) *RouterGroup {
	e := g.engine
	newGroup := &RouterGroup{
		prefix: g.prefix + prefix,
		engine: e,
	}
	e.groups = append(e.groups, newGroup)
	return newGroup
}

// 通过路由分组添加路由信息
func (g *RouterGroup) addRoute(method string, comp string, handler HandleFunc) {
	pattern := g.prefix + comp
	//log.Printf("Route %4s - %s",method,pattern)
	g.engine.router.addRoute(method, pattern, handler)
}

// GET 添加HTTP GET请求的路由
func (g *RouterGroup) GET(pattern string, handler HandleFunc) {
	g.addRoute("GET", pattern, handler)
}

// POST 添加HTTP POST请求的路由
func (g *RouterGroup) POST(pattern string, handler HandleFunc) {
	g.addRoute("POST", pattern, handler)
}

// Use 向路由分组中添加中间件
func (g *RouterGroup) Use(middlewares ...HandleFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}
