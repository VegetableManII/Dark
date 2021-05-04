package main

import (
	"darkorm"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// mysql数据库操作
	/*
		db,err := sql.Open("mysql","root:12345678@tcp(localhost:3306)/myORM")
		defer func() {_ = db.Close()}()
		_,_ = db.Exec("DROP TABLE IF EXISTS User;")
		_,_ = db.Exec("CREATE  TABLE User(Name text);")
		result,err := db.Exec("INSERT INTO User(`Name`) values (?),(?)","Jack","Lucy")
		if err == nil {
			affected,_ := result.RowsAffected()
			log.Println(affected)
		}
		row := db.QueryRow("SELECT Name FROM User ORDER BY Name DESC LIMIT 1")
		var name string
		if err := row.Scan(&name);err == nil {
			log.Println(name)
		}
	*/
	engine, _ := darkorm.NewEngine("mysql", "root:12345678@tcp(localhost:3306)/myORM")
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE  TABLE User(Name text);").Exec()
	// 当执行下条语句时将打印错误日志
	_, _ = s.Raw("CREATE  TABLE User(Name text);").Exec()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?),(?)", "Jack", "Lucy").Exec()
	count, _ := result.RowsAffected()
	fmt.Printf("Exec success, %d affected\n", count)
}
