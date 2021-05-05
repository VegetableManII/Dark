### go学习实战项目

## Dark Web框架

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
    - **Find**：通用查找
      - 不区分结构体对象，会把整张表中的数据全部读取进来

#### Q：

```go
reflect.ValueOf().Type().Len()
reflect.ValueOf().Len()
len()
```

## DarkCache分布式缓存

## DarkRPC框架

