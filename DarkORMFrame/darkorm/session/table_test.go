package session

import (
	"darkorm/dialect"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

type User struct {
	Id   int `darkorm:"PRIMARY KEY"`
	Name string
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	db, err := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/myORM")
	if err != nil {
		t.Fatal(err)
	}
	dial, _ := dialect.GetDialect("mysql")
	s := New(db, dial).Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("Failed to create table User")
	}
}

func TestSession_Model(t *testing.T) {
	db, _ := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/myORM")
	dial, _ := dialect.GetDialect("mysql")
	s := New(db, dial).Model(&User{})
	table := s.RefTable()
	s.Model(&Session{})
	if table.Name != "User" || s.RefTable().Name != "Session" {
		t.Fatal("Failed to change model")
	}
}
