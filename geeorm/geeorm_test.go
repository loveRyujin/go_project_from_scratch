package geeorm

import (
	"testing"

	_ "github.com/loveRyujin/geeorm/dialect"
	_ "github.com/mattn/go-sqlite3"
)

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	if engine.db == nil {
		t.Error("expected db connection")
	}
}

func TestNewEngine_InvalidDriver(t *testing.T) {
	_, err := NewEngine("invalid_driver", ":memory:")
	if err == nil {
		t.Error("expected error for invalid driver")
	}
}

func TestEngine_Session(t *testing.T) {
	engine, err := NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	s := engine.Session()
	if s == nil {
		t.Fatal("expected session, got nil")
	}
	if s.DB() != engine.db {
		t.Error("session should use engine's db")
	}
}

func TestEngine_Integration(t *testing.T) {
	engine, err := NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	defer engine.Close()

	s := engine.Session()

	// 创建表
	s.Raw("CREATE TABLE user (id INTEGER PRIMARY KEY, name TEXT)")
	_, err = s.Exec()
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// 插入
	s.Raw("INSERT INTO user (id, name) VALUES (?, ?)", 1, "test")
	result, err := s.Exec()
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}

	// 查询
	s.Raw("SELECT name FROM user WHERE id = ?", 1)
	row := s.QueryRow()
	var name string
	if err := row.Scan(&name); err != nil {
		t.Fatalf("failed to scan: %v", err)
	}
	if name != "test" {
		t.Errorf("expected 'test', got %s", name)
	}
}
