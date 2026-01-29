package session

import (
	"database/sql"
	"testing"

	"github.com/loveRyujin/geeorm/dialect"
	_ "github.com/mattn/go-sqlite3"
)

var testDialect, _ = dialect.GetDialect("sqlite3")

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	return db
}

func newTestSession(t *testing.T) (*Session, func()) {
	db := newTestDB(t)
	return New(db, testDialect), func() { db.Close() }
}

func TestSession_New(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	s := New(db, testDialect)
	if s == nil {
		t.Fatal("expected session, got nil")
	}
	if s.DB() != db {
		t.Error("expected same db instance")
	}
}

func TestSession_Raw(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Raw("SELECT * FROM users WHERE id = ?", 1)

	if s.sql.String() != "SELECT * FROM users WHERE id = ? " {
		t.Errorf("unexpected sql: %s", s.sql.String())
	}
	if len(s.sqlVars) != 1 || s.sqlVars[0] != 1 {
		t.Errorf("unexpected sqlVars: %v", s.sqlVars)
	}
}

func TestSession_Clear(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	s.Raw("SELECT 1", 1, 2, 3)
	s.Clear()

	if s.sql.String() != "" {
		t.Errorf("expected empty sql, got: %s", s.sql.String())
	}
	if len(s.sqlVars) != 0 {
		t.Errorf("expected empty sqlVars, got: %v", s.sqlVars)
	}
}

func TestSession_Exec(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	// 创建表
	s.Raw("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	_, err := s.Exec()
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// 插入数据
	s.Raw("INSERT INTO test (id, name) VALUES (?, ?)", 1, "tom")
	result, err := s.Exec()
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	affected, _ := result.RowsAffected()
	if affected != 1 {
		t.Errorf("expected 1 row affected, got %d", affected)
	}
}

func TestSession_QueryRow(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	// 准备数据
	s.Raw("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	s.Exec()
	s.Raw("INSERT INTO test (id, name) VALUES (?, ?)", 1, "alice")
	s.Exec()

	// 查询单行
	s.Raw("SELECT name FROM test WHERE id = ?", 1)
	row := s.QueryRow()
	var name string
	if err := row.Scan(&name); err != nil {
		t.Fatalf("failed to scan: %v", err)
	}
	if name != "alice" {
		t.Errorf("expected alice, got %s", name)
	}
}

func TestSession_QueryRows(t *testing.T) {
	s, cleanup := newTestSession(t)
	defer cleanup()

	// 准备数据
	s.Raw("CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
	s.Exec()
	s.Raw("INSERT INTO test (id, name) VALUES (?, ?), (?, ?)", 1, "a", 2, "b")
	s.Exec()

	// 查询多行
	s.Raw("SELECT name FROM test ORDER BY id")
	rows, err := s.QueryRows()
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("unexpected names: %v", names)
	}
}
