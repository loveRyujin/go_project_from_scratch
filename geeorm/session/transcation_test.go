package session

import "testing"

func TestSession_Begin_Commit(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	// 建表
	s.Raw("CREATE TABLE tx_test (id INTEGER PRIMARY KEY, name TEXT)").Exec()

	// 开启事务并提交
	if err := s.Begin(); err != nil {
		t.Fatalf("failed to begin: %v", err)
	}
	s.Raw("INSERT INTO tx_test (id, name) VALUES (?, ?)", 1, "Tom").Exec()
	if err := s.Commit(); err != nil {
		t.Fatalf("failed to commit: %v", err)
	}

	// 提交后数据应存在
	s.Raw("SELECT name FROM tx_test WHERE id = ?", 1)
	row := s.QueryRow()
	var name string
	if err := row.Scan(&name); err != nil {
		t.Fatalf("failed to scan after commit: %v", err)
	}
	if name != "Tom" {
		t.Errorf("expected 'Tom', got '%s'", name)
	}
}

func TestSession_Begin_Rollback(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	// 建表
	s.Raw("CREATE TABLE tx_test (id INTEGER PRIMARY KEY, name TEXT)").Exec()

	// 开启事务并回滚
	if err := s.Begin(); err != nil {
		t.Fatalf("failed to begin: %v", err)
	}
	s.Raw("INSERT INTO tx_test (id, name) VALUES (?, ?)", 1, "Tom").Exec()
	if err := s.Rollback(); err != nil {
		t.Fatalf("failed to rollback: %v", err)
	}

	// 回滚后数据不应存在
	s.Raw("SELECT count(*) FROM tx_test")
	row := s.QueryRow()
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to scan after rollback: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 records after rollback, got %d", count)
	}
}
