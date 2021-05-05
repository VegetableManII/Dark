package clause

import (
	"fmt"
	"reflect"
	"strings"
)

/*
sql语句的拼接
*/
type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}
type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
)

// Set 根据Type类型构造该子句的sql语句
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	sql, vars := generators[name](vars...)
	c.sql[name] = sql
	// sql语句中没有参数则不进行存储
	c.sqlVars[name] = vars
}

// Build 根据传入的Type顺序构造最终的sql语句
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			// 对接口数组进行展开
			// 不展开的话 vars = [[] [Tom] [] [3]]
			// 展开之后 vars = [Tom 3]
			// reflect.ValueOf().Type().Len()
			// reflect.ValueOf().Len()
			// len()
			fmt.Printf("%v\n", reflect.ValueOf(c.sqlVars[order]).Type())
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
