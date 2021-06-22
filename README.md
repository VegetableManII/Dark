### go学习实战项目

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

## DarkORM框架

#### 功能

- 实现数据对象到表结构的映射
- 可以进行插入和查询

#### 功能组件和目录结构

- **log**：负责日志的打印工作，由`log.New()`创建**error**和**info**日志打印工具，分别l两者的暴露`（*Logger).Printf`和`(*Logger).Println`

- **dialect**：数据库方言，根据不同数据库提供不同的**go**数据类型到数据库类型的转换方案。两个接口一个实现数据转换，一个用于检查表存在，在实现不同数据库时要用到。目前只简单实现了**mysql**数据类型的转换，支持不全面，go的所有字符串类型都转换成text/blob类型，没有对char和vchar的支持，无法将string类型作为主键

- **schema**：数据库侧对象，表示一张表（一张表即是一个结构体对象，表中一列即是结构体中一个字段），实现 `Parse()`将接口数据解析成**schema**结构体对象，记录sql语句中作为参数的数据。

  - `Parse()`：返回值为一个**schema**数据结结构体类型（方便切换表的操作）

    ```
    1.reflect.ValueOf()
    获得空接口的值内容，这里获得的是一个指针，指针指向具体的结构体内容，在使用时也是传递结构体对象指针。返回值为Value类型
    2.reflect.Indirect()
    解指针，返回值为Value类型
    3.(Value).Type()
    返回值为Type类型，情景下为 struct 类型
    4.(Type).Name()
    获得结构体的名字，作为schema结构体对象中的字段值
    5.(Type).NumField()->(Type).Field(i)
    返回值为StructField类型，即结构体中的某个字段，调用dialect中的接口进行类型转换作为sql语句中列的类型声明
    6.StructField.Tag.Lookup()
    检查 tag 声明，情景下为该字段（数据库中的列）的约束条件
    ```

- **clause**：表示sql语句的编写规范，实现sql语句的生成规则。

  - 定义了五种sql 关键字：INSERT、VALUES、SELECT、LIMIT、WHERE、ORDERBY

  - 根据不同的关键字调用 `generate()`生成对应sql子句，使用`Build()`对子句进行语句拼接

  - SQL生成器：**generator**

    - ```go
      type generator func(values ...interface{}) (string, []interface{})
      ```

  - 键值映射，根据不同关键字执行不同的`generator`方法

- **session**：

  - 负责一次数据库操作，每一次操作完成都要执行`Clear()`

  - 持有数据库连接、数据库方言、数据库表结构体映射、sql语句 `(strings.Builder)`、sql语句中需要的参数`[]interface{}`。对` (*DB).Query`， `(*DB).QueryRaw`和 `(*DB).Exec`进行了封装。

  - 定义所有对表的操作`CreateTable`，`DropTable`，`HasTable`

    - `Model`方法用于检查当前操作是否进行了切换，操作发生在不同的结构体对象上(表上)，如果是，则需要重新`Parse`解析
    - `RefTable`获得当前正在操作的表

  - 参数记录工具：**record**

    - **Insert**：通用插入
      - 不区分结构体对象，内部会调用`Model`重新解析
    - **Find**：通用查找，参数为结构体数组指针类型，&[]User{}
      - 不区分结构体对象，会把整张表中的数据全部读取进来

    ```go
    1.reflect.ValueOf()
    // 返回值为Value，指针需要解指针，表示的是一个数组的指针
    2.reflect.Indirect()
    // 返回值为Value，解指针，表示具体的数组可以对其进行append操作
    3.(Value).Type().Elem()
    // 返回值为Type
    // 如果不使用Elem得到指针的Type，使用Elem才能得到具体结构体的Type
    4.reflect.New(destType).Elem().Interface()
    // 根据上一步的Type构造对应类型的零值
    // 使用Elem构造的是User类型Value
    // 不使用Elem构造的是*User类型Value
    5.reflect.New(destType).Elem()
    // 获得结构体的Value
    6.(Value).FieldByName(name).Addr().Interface()
    // 获取结构体字段内容，通过Addr转换为指针，进行类型断言转接口
    // 因为进行Scan操作需要使用地址，而且参数类型为接口切片
    7.(*Rows) Scan(dest ...interface{})
    // 读取
    8.(Value).Set(reflect.Append(destSlice, dest))
    // 添加读取的内容
    ```

#### Q：

反射方法对比

|                            Value                             |                             Type                             |
| :----------------------------------------------------------: | :----------------------------------------------------------: |
|              **reflect.ValueOf**返回对象的指针               | **reflect.TypeOf**返回对象的类型，与**(Value).Type()**效果相同 |
| **reflect.Indirect**底层调用**(Value).Elem**解指针，返回对象具体的值，如果接口中数据为基础数据类型则直接返回其保存的实际数值，如果是复合类型，如切片数组，map等返回的是解指针之后的部分，因为复合类型的实现仍包含引用的部分 | **(Type).Elem**只能用在数组、切片、管道、Map或ptr类型上，如结构体数组，则可以返回具体的结构体的Type，例如 **[ ]User{ }**，可以得到User类型的Type通过New构造出的Value可以调用**FieldByName**；对于复合类型必须使用Elem才能得到实际类型的内容，直接使用Type()返回结果是指针Type，使用New构造为空指针nil |
|                                                              |          **reflect.New**根据类型构造对应类型的Value          |
| **(Value).FieldByName**如果接口中的数据为结构体则可以使用该方法，返回一个表示结构体中某个字段的Value |                                                              |
|        **(Value).Addr**返回一个表示调用者Value的地址         |                                                              |
| **(Value).Set**系方法，用于修改调用者Value中的数据内容，可以配合**reflect.Append**实现接口切片的append操作 |                                                              |

```go
reflect.ValueOf().Type().Len()
reflect.ValueOf().Len()
len()
```

## DarkCache分布式缓存

#### 功能

- LRU缓存淘汰策略


- 单机并发缓存

- HTTP服务端
- 一致性哈希及虚拟节点

#### 功能组件和目录结构

- **lru**：
  - 主要成员
    - 双向链表：`container/list`
    - Hash表：`map`
    - OnEvicted(函数指针成员)：当某条记录移除时触发回调函数
  - 主要方法
    - **Get**：通过key获取`*list.Element`，并将其移动到首部，通过类型断言将`*list.Element`转换为节点类型
    - **RemoveOldest**：获取最后一个元素删除（双向链表和hash表中都要删除），更新使用内存大小，调用回调函数如果有的话
    - **Add**：如果已存在则移动到首部，更新使用内存大小,更新hash表对应值；如果不存在，创建节点并插入到首部在hash表中添加并更新使用内存大小，如果内存超过最大限制则调用`RemoveOldest`，直到小于最大限制，最大限制设置为0表示不限制缓存的内存占用大小
- **byteview**：只读数据结构，表示缓存的值，底层为byte数组，提供**ByteSlice**返回缓存数据的拷贝
- **cache**：通过互斥锁实现单机并发访问，对lru进行了加一层锁的封装，`add`和`get`方法对数据的表述均为`byteview`格式
- **Group**：负责与用户交互以及控制缓存值的存储和获取的流程
  - 主要成员：
    - `name`区分不同缓存空间，即缓存命名空间
    - `Getter`回调的注册可以是函数类型也可以结构体类型，当缓存未命中时回调
    - `cache`缓存
  - 主要方法：
    - **Get**：返回类型为`ByteView`，封装`cache`的`get`方法，如果没有缓存则回调加载源数据
    - **getLocally**：
- **Getter**接口：回调函数，当缓存不存在时调用用户提供的回调函数去获取数据或做其他处理。持有`Get`方法
  - `Get(key string) ([]byte, error)`方法被定义为`GetterFunc`类型
  - `GetterFunc`实现了`Get`方法，这种函数成为接口函数，在调用时既可以传入结构体也可以传入这种类型的函数作为参数

## DarkRPC框架

#### 功能

- 消息编解码（序列化与反序列化）

#### 功能组件和目录结构

（客户端发送的请求包括服务名，方法名和参数，服务端的响应包括错误`error`，返回值`replay`）

- **codec**：

  - 定义`Header`，请求和响应中的参数和返回值抽象为`body`，剩余的信息组成`header`（请求服务名方法名，序号，错误）

  - **Codec**接口：

    - 接口方法

      ```go
      	io.Closer
      	ReadHeader(*Header) error
      	ReadBody(interface{}) error
      	Write(*Header, interface{}) error
      ```

    - 实现不同的`Codec`实例，根据不同的类型返回不同的构造函数（提供`Gob`编码和`Json`编码两种）

    - **GobCodec**：持有`gob`的编解码工具以及一个`bufio.Writer`（通过**socket**链接来创建缓冲输出），实现`Codec`接口

- **server**：默认消息编码格式为`Gob`编码，协商时使用`json`编码`option`，后续交流通过`option`中制定的编码格式。持有默认**server**实例使用默认**option**。`Accept`接收连接后`ServeConn`处理

  - **Option**：用于协商消息的编码格式
  - **ServeConn**
    - 创建`json.NewDecoder`解码后获得`option`，根据其**CodecType**类型选择对应的**Codec**类型
  - **serveCodec**
    - 读取请求**readRequest**
    - 处理请求**handleRequest**：并发处理请求使用`sync.Mutex`和`sync.WaitGroup`控制
    - 回复请求**sendResponse**：回复请求需要挨个发送使用`sync.Mutex`来保证