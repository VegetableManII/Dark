package dark

import (
	"log"
	"net/http"
	"strings"
)

// 路由
type router struct {
	roots map[string]*node
	// eg. handlers['GET'] handlers['POST']
	handlers map[string]HandleFunc
	// eg. handlers['GET-/p/:lang/doc'] handlers['POST-/p/book']
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandleFunc),
	}
}

// 解析pattern，分割字符串
// 以 "/" 分割出每一个节点的 part内容
// 解析规则遇到 "*" 结束之后内容的解析
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")
	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			// 如果路由路径中出现通配符则先处理通配符前面是路由
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

/*
注册路由 完整请求路径 = 请求方法 + 匹配规则（带 "/"）
参数：请求方法，匹配规则，响应函数
解析匹配规则，分割字符串得到每一个节点的part
根据请求方法获得根节点node
根据匹配规则建立前缀树
根据完整的请求路径
*/
func (r *router) addRoute(method string, pattern string, handler HandleFunc) {
	parts := parsePattern(pattern)
	log.Printf("Route %4s - %s", method, pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

/*
路由匹配
参数：请求方法，请求路径
params定义了请求路径中的参数集合
解析路径成 part 的数组，依次去前缀树中查找返回查找到的节点
根据请求方法获得前缀树的根节点
根据返回的节点中保存的 pattern 解析出各节点的 part 内容
如果某个节点为通配符节点，即包含 ":" 或 "*" 需要进行处理去掉通配符
并把内容替换为用户请求的具体内容
eg.
	path：/p/golang/doc
	searchParts：[p],[golang],[doc]
	n.pattern: [p],[:lang],[doc]
*/
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string, 0)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			}
			if part[0] == '*' {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}
	return nil, nil
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		// 从路由匹配到的Handler添加到中间件然后Next执行
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND %s\n", c.Path)
		})
	}
	c.Next()
}
