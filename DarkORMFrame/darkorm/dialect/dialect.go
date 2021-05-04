package dialect

import "reflect"

// Dialect实现了go数据类型和数据库数据类型的转换
// 提供解接口根据不同数据库进行具体的实现
type Dialect interface {
	// 数据转换
	DataTypeOf(typ reflect.Value) string
	// 返回值为某个表是否存在，参数为表名
	TableExistSQL(table string) (string, []interface{})
}

var dialectMap = map[string]Dialect{}

func RegisterDialect(name string, dialect Dialect) {
	dialectMap[name] = dialect
}
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectMap[name]
	return
}
