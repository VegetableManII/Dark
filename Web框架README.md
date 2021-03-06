## Dark Web框架

#### 功能

- 前缀树路由
- 通过路由划分实现的访问分组控制
- 可拓展中间件
- 支持模板
- 错误恢复

#### 功能组件和目录结构

- **engine**：定义框架中的一些数据结构，如`Engine`，`HandleFunc`，`RouterGroup`等。

  ```go
  type Engine struct {
     *RouterGroup
     // 继承RouterGroup直接使用RouterGroup中的一些方法
     router *router
     groups []*RouterGroup
     // 加载所有模板到内存
     htmlTemplates *template.Template
     // 所有自定模板渲染函数
     funcMap template.FuncMap
  }
  // HandleFunc 定义请求处理句柄
  type HandleFunc func(c *Context)
  // 路由分组
  type RouterGroup struct {
  	prefix      string
  	middlewares []HandleFunc // 不同分组使用不同的中间件
  	engine      *Engine      // 所有分组共享一个engin实例
  }
  ```

  - **Engine**相关方法
    - 调用`router`中的`addRoute`方法添加路由实现GET，POST等请求的处理
    - 调用`RouterGroup`中的`Static`方法处理静态资源
    - 调用`RouterGroup`中的`Use`添加中间件，实现最顶层的中间件注册而不是在某一条路径中运行的中间件
    - 实现 `Handler` 的 `ServeHTTP` 接口，所有的HTTP请求都会通过该函数进入处理，调用`router`中的`handle`方法传递新创建的`Context`然后进行处理
    - 提供`SetFunMap`支持设置自定义模板渲染、提供`LoadHTMLGlob`加载模板
  - **RouterGroup**相关方法
    - **addRoute**：拼接路由分组的前缀，通过调用`engine`中的成员`router`添加路由
    - **Use**：向路由分组中添加中间件，即添加`HandleFunc`函数到`RouterGroup`中
    - **createStaticHandler**：创建静态资源处理函数
    - **Static**：注册静态资源
  - **HandleFunc**：参数为 ***Context** 的函数

- **context**：实现了JSON、HTML、String以及Data格式数据流的输出，

  - 通过`(ResponseWriter).WriteHeader`和`(ResponseWriter).Header().Set`设置响应头部信息
  - 通过`(*Encoder)Encode()`编码并输出JSON数据流
  - 通过`(ResponseWriter).Write`输出二进制数据流
  - 通过`(*Templeate).ExecuteTemplate`输出HTML模板文件

  ```go
  type Context struct {
     Writer http.ResponseWriter
     Req    *http.Request
     // 请求信息
     Path   string
     Method string
     // 响应信息
     StatusCode int
     // 路由参数
     Params map[string]string
     // 中间件
     handlers []HandleFunc
     index    int
    // 用于Context中handlers中间件的全部执行
     engin    *Engine
  }
  ```

  - 封装`(*Request).FormValue`和`Request.URL.Query().Get()`来获取POST和GET的请求参数
  - **Next**：遍历`handlers`执行中间件方法，在`router`中赋值

- **router**：

  ```go
  type router struct {
    // eg. handlers['GET-/p/:lang/doc'] handlers['POST-/p/book']
     roots map[string]*node
     // node 为前缀树上的某个节点 
     handlers map[string]HandleFunc
    // 保存中间件方法，便于给context赋值,查找是通过完整路由(方法-路径)作为键来实现
  }
  ```

  - **addRoute**：解析请求路径得到字符串数组（以"/"分割，遇到"*"之后的字段直接忽略）。方法和路径作为键前缀树节点作为值，如果没有则创建一个新的节点，如果有则根据路径进行插入。同时`handlers`中对应存储中间件的方法

  - **getRoute**：解析请求路径得到字符串数组，获得前以请求方法GET或POST作为根节点的前缀树的根节点，根据字符串数组进行查找，创建string=>string的map，将的到的节点再次解析分析路径将其中的参数保存在map中。

    规定，使用*****代表通配符其后的内容都作为值，使用 **:** 作为通配符只把 `../:example/..`example字段作为值进行替换。例如，某节点保存路径为`GET-/p/:lang/doc`，其中`lang`会作为键，而用户请求的`GET-/p/Java/doc`，其中`Java`作为值进行替换。不能使用诸如**/assets/***这样的路由，使用*****其后必须紧跟一种系统关键字，如`/assets/*file`可以使用`file`系统关键字来识别文件

  - **handle**：通过`getRoute`得到路由参数map赋予context并把对应的中间件`handlers`赋值，如果路由为空则向context中添加`HandleFunc`输出404响应，然后调用`(*Context).Next`

- **logger**：返回一个`HandleFunc`，其内容是记录当前时间，调用`(*Context).Next`，然后打印日志

  context的执行流程为一个递归然后回溯的过程

- **recovery**：负责引擎的持久运行，实现了错误恢复

  - **trace**：

    - `runtime.Callers(3,pcs[:])`获得发生`panic`的代码段，第一个参数表示捕获的递归跳转层次，**0**表示本身`Callers`本身，**1**表示调用Callers的`trace`，**2**表示调用`trace`的`Recovery`，3表示用户代码调用`Recovery`的函数。
    - `runtime.FuncForPC`，根据`uintptr`指针获取到对应的`Func`，`Func`通过指针作为参数调用`(*Func)FileLine(uintptr)`获得文件名和行数

  - **Recovery**：提供给用户的错误恢复方法，可以在可能发生错误的地方调用

    - 返回一个`HandleFunc`，函数中使用`defer`和`recover`实现恢复，先继续执行`(*Context).Next`让系统继续运行，然后调用`defer`恢复，当`recover`返回非nil时，`Context`的当前`HandleFunc`发生错误不影响后面的继续执行

    ```go
    return func(c *Context) {
       defer func() {
          if err := recover();err != nil {
             message := fmt.Sprintf("%s",err)
             log.Printf("%s\n",trace(message))
             c.Fail(http.StatusInternalServerError,"Internal Server Error")
          }
       }()
       c.Next()
    }
    ```

- **trie**：实现了路由路径的前缀树结构

  ```go
  type node struct {
     pattern  string  // 待匹配路由，在叶子节点该字段才有实际值
     part     string  // 路由中的部分内容，即请求路径中相邻两个 “/” 之间的字段
     children []*node // 子节点，
     isWild   bool    // 是否模糊匹配，提供两种路由参数的通配符 ：和 *
  }
  ```

  - **insert**：
    - 递归终止条件，通过参数中的字符串数组判断树的查询高度，如果当前高度和字符串长度一致则将当前节点的pattern设置为全路径，当前节点成为叶子节点；
    - 如果不是叶子节点继续进行查询，如果无法找到匹配的子节点则创建一个子节点再进行一次递归对该叶子节点的`pattern`进行赋值
  - **search**：
    - 递归终止条件，如果查询高度等于字符串数组长度或者当前节点的part字段含有通配符，则进行pattern字段判断，如果pattern为空说明非叶子节点，返回nil，匹配失败，只会返回叶子节点中的pattern字段。
    - 进行层次查询遍历每一层的所有孩子节点然后挨个从孩子节点中查询目标，找不到返回nil
  - **matchChild**：匹配某个节点，条件：part字段相等或不进行精准匹配 `isWild`字段为true
  - **matchChildren**：匹配某个节点的所有孩子，条件同上

#### Q：

```
// 前缀树的判断终止条件，在router中getRoute时对通配符之后的操作都进行的自动赋值
// 这里只需要查找到第一次出现通配符的位置即可
if len(parts) == height || strings.HasPrefix(n.part, "*") || strings.HasPrefix(n.part, ":") {
   if n.pattern == "" {
      return nil
   }
   return n
}
```

```go
reflect.ValueOf().Type().Len()
reflect.ValueOf().Len()
len()
```
