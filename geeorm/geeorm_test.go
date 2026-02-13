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

// newBareEngine 创建一个不带预建表的 Engine。
func newBareEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	return engine
}

// newTestEngine 创建 Engine 并预建 User 表。
func newTestEngine(t *testing.T) *Engine {
	t.Helper()
	engine := newBareEngine(t)
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

func TestMigrate_CreateTable(t *testing.T) {
	engine := newBareEngine(t)
	defer engine.Close()

	// 表不存在时，Migrate 应自动创建
	if err := engine.Migrate(&User{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	s := engine.Session()
	if !s.Model(&User{}).HasTable() {
		t.Error("expected table to be created")
	}
}

func TestMigrate_AddColumn(t *testing.T) {
	engine := newBareEngine(t)
	defer engine.Close()

	// 先用 raw SQL 建一张只有 Name 列的 User 表
	s := engine.Session()
	_, _ = s.Raw("CREATE TABLE User (Name text PRIMARY KEY);").Exec()
	_, _ = s.Raw("INSERT INTO User (Name) VALUES (?)", "Tom").Exec()

	// Migrate 应当添加 Age 列
	if err := engine.Migrate(&User{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// 验证新列存在（已有行的 Age 为 NULL，通过列元数据确认）
	s = engine.Session()
	rows, err := s.Raw("SELECT * FROM User LIMIT 1").QueryRows()
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	cols, _ := rows.Columns()
	_ = rows.Close() // 必须关闭，否则 :memory: 连接池会打开新连接看到空库
	hasAge := false
	for _, col := range cols {
		if col == "Age" {
			hasAge = true
		}
	}
	if !hasAge {
		t.Error("expected Age column to be added")
	}

	// 插入新行验证列可正常使用
	s = engine.Session()
	s.Model(&User{})
	if _, err := s.Insert(&User{"Sam", 25}); err != nil {
		t.Fatalf("insert after migrate failed: %v", err)
	}
}

func TestMigrate_DeleteColumn(t *testing.T) {
	engine := newBareEngine(t)
	defer engine.Close()

	// 先用 raw SQL 建一张多出 Email 列的 User 表
	s := engine.Session()
	_, _ = s.Raw("CREATE TABLE User (Name text PRIMARY KEY, Age integer, Email text);").Exec()
	_, _ = s.Raw("INSERT INTO User (Name, Age, Email) VALUES (?, ?, ?)", "Tom", 18, "tom@test.com").Exec()

	// Migrate 应当删除 Email 列
	if err := engine.Migrate(&User{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// 验证数据保留且 Email 列已移除
	s = engine.Session()
	var users []User
	if err := s.Model(&User{}).Find(&users); err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if len(users) != 1 || users[0].Name != "Tom" || users[0].Age != 18 {
		t.Errorf("expected [{Tom 18}], got %v", users)
	}

	// 确认 Email 列已不在表中
	rows, err := s.Raw("SELECT * FROM User LIMIT 1").QueryRows()
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	cols, _ := rows.Columns()
	for _, col := range cols {
		if col == "Email" {
			t.Error("expected Email column to be removed")
		}
	}
}

func TestMigrate_NoChange(t *testing.T) {
	engine := newTestEngine(t)
	defer engine.Close()

	// 插入一条数据
	s := engine.Session()
	s.Model(&User{})
	_, _ = s.Insert(&User{"Tom", 18})

	// Migrate 无变更应正常返回
	if err := engine.Migrate(&User{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	// 验证数据不受影响
	s = engine.Session()
	s.Model(&User{})
	count, _ := s.Count()
	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}
}
