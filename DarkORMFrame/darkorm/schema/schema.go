package schema

import (
	"darkorm/dialect"
	"go/ast"
	"reflect"
)

// Field 表示数据库中的一列
type Field struct {
	Name string //字段名
	Type string // 类型
	Tag  string // 约束条件
}

// Schema 表示数据库中的一张表
type Schema struct {
	Model      interface{}
	Name       string
	Fields     []*Field          // 保存所有字段
	FieldNames []string          // 保存所有字段的名称
	fieldMap   map[string]*Field // 方便快速找到对应字段 时间复杂度 O(1)
}

func (sche *Schema) GetField(name string) *Field {
	return sche.fieldMap[name]
}

// Parse 将任意对象解析成Schema实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// TypeOf 和 ValueOf 返回参数的类型和值
	// 通过ValueOf获取到类型的值，返回值Value(保存着指针和底层类型)，创建对应类型的内容保存在dest中
	// 通过Indirect根据Value结构体对象获取到其中的指针所保存的具体内容，返回值为Value
	// 通过Value结构体对象的方法Type获得成员变量typ表示的底层类型——结构体类型

	// 如果使用reflect.ValueOf(dest).Type()直接进行反射解析得到的是结构体指针类型
	// reflect.ValueOf(dest).Elem().Type() 和 reflect.Indirect 是等价的，indirect 对参数为nil是进行了保护
	//modeType := reflect.ValueOf(dest).Type()
	//fmt.Printf("modeType is %+v\n",modeType)
	//modeType = reflect.ValueOf(dest).Elem().Type()
	//fmt.Printf("modeType is %+v\n",modeType)
	modeType := reflect.Indirect(reflect.ValueOf(dest)).Type()

	schema := &Schema{
		Model:    dest,
		Name:     modeType.Name(), // modeType底层是结构体类型则通过Name方法获得结构体名称
		fieldMap: make(map[string]*Field),
	}
	// NumField能够获得结构体中的字段数
	for i := 0; i < modeType.NumField(); i++ {
		p := modeType.Field(i) // 依次获取字段
		// 检查字段是否是匿名字段以及是否是公开暴露的字段
		if !p.Anonymous && ast.IsExported(p.Name) {

			//fmt.Printf("p is %+v\n",reflect.New(p.Type))

			field := &Field{
				Name: p.Name, // 字段的名字
				// 使用reflect.New构造出来的反射类型也是指针类型需要调用reflect.Indirect来解指针
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))), //字段的类型
			}
			// 结构体的声明中使用 tag 工具设置约束条件，如主键等
			/*
				ype User struct {
					Name string `darkorm:"PRIMARY KEY"`
					Age  int
				}
			*/
			if v, ok := p.Tag.Lookup("darkorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

func (sche *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range sche.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
