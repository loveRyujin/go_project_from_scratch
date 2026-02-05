package session

import (
	"testing"
)

func testRecordInit(t *testing.T) (*Session, func()) {
	t.Helper()
	s, cleanup := newTestSession(t)
	s.Model(&User{})
	_ = s.DropTable()
	err := s.CreateTable()
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return s, cleanup
}

func TestSession_Insert(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 测试插入单条记录
	affected, err := s.Insert(&User{"Tom", 18})
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}
}

func TestSession_Insert_Multiple(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 测试批量插入多条记录
	user1 := &User{"Tom", 18}
	user2 := &User{"Sam", 25}
	user3 := &User{"Jack", 30}

	affected, err := s.Insert(user1, user2, user3)
	if err != nil {
		t.Fatalf("failed to insert multiple: %v", err)
	}
	if affected != 3 {
		t.Errorf("expected 3 rows affected, got %d", affected)
	}
}

func TestSession_Find(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 先插入测试数据
	_, _ = s.Insert(&User{"Tom", 18}, &User{"Sam", 25})

	// 查询所有记录
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestSession_Find_VerifyData(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 插入已知数据
	_, _ = s.Insert(&User{"Alice", 20})

	// 查询并验证数据正确性
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Name != "Alice" || users[0].Age != 20 {
		t.Errorf("expected {Alice, 20}, got {%s, %d}", users[0].Name, users[0].Age)
	}
}

func TestSession_Find_Empty(t *testing.T) {
	s, cleanup := testRecordInit(t)
	defer cleanup()

	// 空表查询
	var users []User
	err := s.Find(&users)
	if err != nil {
		t.Fatalf("failed to find from empty table: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected 0 users, got %d", len(users))
	}
}
