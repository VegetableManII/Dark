package dark

import (
	"html/template"
	"net/http"
	"path"
	"strings"
)

// 路由分组
type RouterGroup struct {
	prefix      string
	middlewares []HandleFunc //支持中间件,不同前缀分组可使用不同的中间件
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
	// 加载所有模板到内存
	htmlTemplates *template.Template
	// 所有自定模板渲染函数
	funcMap template.FuncMap
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

// SetFuncMap 设置自定义模板渲染函数
func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

// LoadHTMLGlob 加载模板
func (e *Engine) LoadHTMLGlob(pattern string) {
	e.htmlTemplates = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

// ServeHTTP接口，所有的HTTP请求都会通过该函数进入处理
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
	c.engin = e
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

// 创建处理静态请求的Handler
func (g *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandleFunc {
	absolutePath := path.Join(g.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("file")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// 处理静态资源
func (g *RouterGroup) Static(relativePath string, root string) {
	handler := g.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*file")
	// 注册 GET 方法
	g.GET(urlPattern, handler)
}
