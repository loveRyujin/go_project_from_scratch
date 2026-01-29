package session

import (
	"testing"

	"github.com/loveRyujin/geeorm/dialect"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func TestSession_Model(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Model(&User{})
	table := s.RefTable()

	if table == nil {
		t.Fatal("expected table, got nil")
	}
	if table.Name != "User" {
		t.Errorf("expected table name 'User', got %s", table.Name)
	}
	if len(table.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(table.Fields))
	}
}

func TestSession_CreateTable(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Model(&User{})
	_ = s.DropTable()
	err := s.CreateTable()
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	if !s.HasTable() {
		t.Fatal("Failed to create table User")
	}
}

func TestSession_DropTable(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Model(&User{})
	_ = s.CreateTable()
	if !s.HasTable() {
		t.Fatal("expected table to exist before drop")
	}

	err := s.DropTable()
	if err != nil {
		t.Fatalf("failed to drop table: %v", err)
	}
	if s.HasTable() {
		t.Error("expected table to be dropped")
	}
}

func TestSession_HasTable(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Model(&User{})

	// 表不存在
	if s.HasTable() {
		t.Error("expected table not to exist initially")
	}

	// 创建表后应该存在
	s.CreateTable()
	if !s.HasTable() {
		t.Error("expected table to exist after creation")
	}
}

func TestSession_CreateTable_WithTag(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Model(&User{})
	s.CreateTable()

	// 验证 PRIMARY KEY 约束生效 - 尝试插入重复主键应该失败
	s.Raw("INSERT INTO User (Name, Age) VALUES (?, ?)", "tom", 18)
	_, err := s.Exec()
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}

	s.Raw("INSERT INTO User (Name, Age) VALUES (?, ?)", "tom", 20)
	_, err = s.Exec()
	if err == nil {
		t.Error("expected error for duplicate primary key")
	}
}

func init() {
	// 确保 dialect 已注册
	if _, ok := dialect.GetDialect("sqlite3"); !ok {
		panic("sqlite3 dialect not registered")
	}
}
