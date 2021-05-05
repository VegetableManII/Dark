package clause

import (
	"reflect"
	"testing"
)

func testSelect(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name=?", "Tom")
	clause.Set(ORDERBY, "Age ASC")
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User WHERE Name=? ORDER BY Age ASC LIMIT ?" {
		t.Fatal("failed to build SQL")
	}
	// 这里vars的返回值类型为 三层interface数组的嵌套
	// 第一层数据类型为 interface 数组 每一个元素类型为 interface 数组
	// 第二层数据类型为 interface 数组 每一个元素类型为 interface | some-type
	// 第三层数据类型为 带有实际数据的interface类型 interface | string 和 interface | int
	// []interface{}{    []interface{}{"Tom"}   ,    []interface{}{3}       }
	if !reflect.DeepEqual(vars, []interface{}{"Tom", 3}) {
		t.Fatal("failed to build SQLVars")
	}
}

func TestClause_Build(t *testing.T) {
	t.Run("select", func(t *testing.T) {
		testSelect(t)
	})
}
