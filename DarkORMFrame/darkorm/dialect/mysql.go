package dialect

/*
对mysql支持不全面，go的所有字符串类型都转换成text/blob类型
没有对char和vchar的支持，无法将go的string类型作为主键
*/
import (
	"fmt"
	"reflect"
	"time"
)

type mysql struct{}

// 用来检验mysql是否实现了Dialect的接口
// 如果没有实现该语句错误
var _ Dialect = (*mysql)(nil)

func init() {
	RegisterDialect("mysql", &mysql{})
}

func (mq *mysql) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int8, reflect.Uint8:
		return "tinyint"
	case reflect.Int16, reflect.Uint16:
		return "smallint"
		// go 没有三字节数据类型 所以无法映射 mediumint 类型
	case reflect.Int32, reflect.Uint32, reflect.Int, reflect.Uint, reflect.Uintptr:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32:
		return "float"
	case reflect.Float64:
		return "double"
	case reflect.String:
		// text 大小为 0~65535字节，而go的String类型可能超过该大小
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

func (mq *mysql) TableExistSQL(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME=?", args
}
