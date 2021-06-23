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