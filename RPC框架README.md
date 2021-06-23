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