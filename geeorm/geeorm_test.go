package geeorm

import (
	"errors"
	"testing"

	_ "github.com/loveRyujin/geeorm/dialect"
	"github.com/loveRyujin/geeorm/session"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	s := engine.Session()
	s.Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	return engine
}

func TestNewEngine(t *testing.T) {
	engine := newTestEngine(t)
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
	engine := newTestEngine(t)
	defer engine.Close()

	s := engine.Session()
	if s == nil {
		t.Fatal("expected session, got nil")
	}
}

func TestEngine_Integration(t *testing.T) {
	engine := newTestEngine(t)
	defer engine.Close()

	s := engine.Session()
	s.Model(&User{})

	// 插入
	affected, err := s.Insert(&User{"Tom", 18})
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}

	// 查询
	var user User
	if err := s.Where("Name = ?", "Tom").First(&user); err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if user.Name != "Tom" || user.Age != 18 {
		t.Errorf("expected {Tom, 18}, got {%s, %d}", user.Name, user.Age)
	}
}

func TestTranscation_Commit(t *testing.T) {
	engine := newTestEngine(t)
	defer engine.Close()

	// 事务正常提交
	_, err := engine.Transcation(func(s *session.Session) (any, error) {
		s.Model(&User{})
		_, err := s.Insert(&User{"Tom", 18})
		return nil, err
	})
	if err != nil {
		t.Fatalf("transaction commit failed: %v", err)
	}

	// 验证数据已持久化
	s := engine.Session()
	s.Model(&User{})
	count, _ := s.Count()
	if count != 1 {
		t.Errorf("expected 1 record after commit, got %d", count)
	}
}

func TestTranscation_Rollback(t *testing.T) {
	engine := newTestEngine(t)
	defer engine.Close()

	// 事务返回错误，应自动回滚
	_, err := engine.Transcation(func(s *session.Session) (any, error) {
		s.Model(&User{})
		s.Insert(&User{"Tom", 18})
		return nil, errors.New("simulate error")
	})
	if err == nil {
		t.Fatal("expected error from transaction")
	}

	// 验证数据已回滚
	s := engine.Session()
	s.Model(&User{})
	count, _ := s.Count()
	if count != 0 {
		t.Errorf("expected 0 records after rollback, got %d", count)
	}
}

func TestTranscation_Panic(t *testing.T) {
	engine := newTestEngine(t)
	defer engine.Close()

	// 事务中 panic，应回滚并重新 panic
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic to be re-raised")
		}

		// 验证数据已回滚
		s := engine.Session()
		s.Model(&User{})
		count, _ := s.Count()
		if count != 0 {
			t.Errorf("expected 0 records after panic rollback, got %d", count)
		}
	}()

	engine.Transcation(func(s *session.Session) (any, error) {
		s.Model(&User{})
		s.Insert(&User{"Tom", 18})
		panic("simulate panic")
	})
}
