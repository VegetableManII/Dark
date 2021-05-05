package session

import (
	"darkorm/dialect"
	"database/sql"
	"testing"
)

var (
	user1 = &User{1, "Tom", 18}
	user2 = &User{2, "Jack", 19}
	user3 = &User{3, "Jerry", 20}
)

func testRecordInit(t *testing.T) *Session {
	t.Helper()
	db, _ := sql.Open("mysql", "root:12345678@tcp(localhost:3306)/myORM")
	dial, _ := dialect.GetDialect("mysql")
	s := New(db, dial).Model(&User{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("failed init test records")
	}
	return s
}

func TestSession_Insert(t *testing.T) {
	s := testRecordInit(t)
	affected, err := s.Insert(user3)
	if err != nil || affected != 1 {
		t.Fatal("failed to create record")
	}
}

func TestSession_Find(t *testing.T) {
	s := testRecordInit(t)
	var users []User
	if err := s.Find(&users); err != nil || len(users) != 2 {
		t.Fatal("failed to query all")
	}
}
