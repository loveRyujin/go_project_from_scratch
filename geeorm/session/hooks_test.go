package session

import (
	"errors"
	"testing"
)

// Account 实现了 BeforeInsert 和 AfterQuery 钩子
type Account struct {
	Name     string `geeorm:"PRIMARY KEY"`
	Password string
}

// BeforeInsert 插入前自动处理密码（模拟加密）
func (a *Account) BeforeInsert(s *Session) error {
	a.Password = "*" + a.Password + "*"
	return nil
}

// AfterQuery 查询后自动处理密码（模拟解密）
func (a *Account) AfterQuery(s *Session) error {
	a.Password = a.Password[1 : len(a.Password)-1]
	return nil
}

func testHookInit(t *testing.T) (*Session, func()) {
	t.Helper()
	s, cleanup := newTestSession(t)
	s.Model(&Account{})
	_ = s.DropTable()
	if err := s.CreateTable(); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return s, cleanup
}

func TestHook_BeforeInsert(t *testing.T) {
	s, cleanup := testHookInit(t)
	defer cleanup()

	// BeforeInsert 应该把 Password 包装为 *secret*
	_, err := s.Insert(&Account{"Tom", "secret"})
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// 验证数据库中存储的是处理后的密码
	var accounts []Account
	_ = s.Find(&accounts) // AfterQuery 会解密回 secret
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}
	if accounts[0].Password != "secret" {
		t.Errorf("expected password 'secret' after AfterQuery, got '%s'", accounts[0].Password)
	}
}

func TestHook_AfterQuery(t *testing.T) {
	s, cleanup := testHookInit(t)
	defer cleanup()

	// 手动插入已"加密"的数据
	s.Raw("INSERT INTO Account (Name, Password) VALUES (?, ?)", "Tom", "*raw*").Exec()

	// AfterQuery 应把 *raw* 解密为 raw
	var accounts []Account
	_ = s.Find(&accounts)
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}
	if accounts[0].Password != "raw" {
		t.Errorf("expected password 'raw' after AfterQuery, got '%s'", accounts[0].Password)
	}
}

// FailModel 实现一个会失败的 BeforeInsert 钩子
type FailModel struct {
	Name string `geeorm:"PRIMARY KEY"`
}

func (f *FailModel) BeforeInsert(s *Session) error {
	return errors.New("insert blocked by hook")
}

func TestHook_BeforeInsert_Error(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Model(&FailModel{})
	_ = s.DropTable()
	_ = s.CreateTable()

	// BeforeInsert 返回错误，插入应被阻止
	_, err := s.Insert(&FailModel{"Tom"})
	if err == nil {
		t.Fatal("expected error from BeforeInsert hook")
	}
	if err.Error() != "insert blocked by hook" {
		t.Errorf("expected 'insert blocked by hook', got '%s'", err.Error())
	}

	// 确认没有数据被插入
	count, _ := s.Count()
	if count != 0 {
		t.Errorf("expected 0 records after blocked insert, got %d", count)
	}
}

// NoHookUser 没有实现任何钩子，验证静默跳过
func TestHook_NoHook(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// User 没有实现任何 hook，Insert/Find 应正常工作
	_, err := s.Insert(&User{"Tom", 18})
	if err != nil {
		t.Fatalf("insert without hook failed: %v", err)
	}

	var users []User
	err = s.Find(&users)
	if err != nil {
		t.Fatalf("find without hook failed: %v", err)
	}
	if len(users) != 1 || users[0].Name != "Tom" {
		t.Errorf("expected [{Tom 18}], got %v", users)
	}
}
