package clause

import (
	"reflect"
	"testing"
)

func testSelect(t *testing.T) {
	var clause Clause
	clause.Set(LIMIT, 3)
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(WHERE, "Name = ?", "Tom")
	clause.Set(ORDERBY, "Age", false)
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User WHERE Name = ? ORDER BY Age LIMIT ?" {
		t.Fatalf("failed to build SQL, got: %s", sql)
	}
	if !reflect.DeepEqual(vars, []any{"Tom", 3}) {
		t.Fatal("failed to build SQLVars")
	}
}

func testInsert(t *testing.T) {
	var clause Clause
	clause.Set(INSERT, "User", []string{"Name", "Age"})
	clause.Set(VALUES, []any{"Tom", 18}, []any{"Sam", 25})
	sql, vars := clause.Build(INSERT, VALUES)
	t.Log(sql, vars)
	if sql != "INSERT INTO User (Name,Age) VALUES (?, ?), (?, ?)" {
		t.Fatalf("failed to build SQL, got: %s", sql)
	}
	if !reflect.DeepEqual(vars, []any{"Tom", 18, "Sam", 25}) {
		t.Fatalf("failed to build SQLVars, got: %v", vars)
	}
}

func testUpdate(t *testing.T) {
	var clause Clause
	clause.Set(UPDATE, "User", map[string]any{"Age": 30})
	clause.Set(WHERE, "Name = ?", "Tom")
	sql, vars := clause.Build(UPDATE, WHERE)
	t.Log(sql, vars)
	if sql != "UPDATE User SET Age = ? WHERE Name = ?" {
		t.Fatalf("failed to build SQL, got: %s", sql)
	}
	if !reflect.DeepEqual(vars, []any{30, "Tom"}) {
		t.Fatalf("failed to build SQLVars, got: %v", vars)
	}
}

func testDelete(t *testing.T) {
	var clause Clause
	clause.Set(DELETE, "User")
	clause.Set(WHERE, "Name = ?", "Tom")
	sql, vars := clause.Build(DELETE, WHERE)
	t.Log(sql, vars)
	if sql != "DELETE FROM User WHERE Name = ?" {
		t.Fatalf("failed to build SQL, got: %s", sql)
	}
	if !reflect.DeepEqual(vars, []any{"Tom"}) {
		t.Fatalf("failed to build SQLVars, got: %v", vars)
	}
}

func testCount(t *testing.T) {
	var clause Clause
	clause.Set(COUNT, "User")
	sql, vars := clause.Build(COUNT)
	t.Log(sql, vars)
	if sql != "SELECT count(*) FROM User" {
		t.Fatalf("failed to build SQL, got: %s", sql)
	}
	if len(vars) != 0 {
		t.Fatalf("expected no vars, got: %v", vars)
	}
}

func testOrderByDesc(t *testing.T) {
	var clause Clause
	clause.Set(SELECT, "User", []string{"*"})
	clause.Set(ORDERBY, "Age", true)
	sql, vars := clause.Build(SELECT, ORDERBY)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User ORDER BY Age DESC" {
		t.Fatalf("failed to build SQL, got: %s", sql)
	}
	if len(vars) != 0 {
		t.Fatalf("expected no vars, got: %v", vars)
	}
}

func TestClause_Build(t *testing.T) {
	t.Run("select", func(t *testing.T) {
		testSelect(t)
	})
	t.Run("insert", func(t *testing.T) {
		testInsert(t)
	})
	t.Run("update", func(t *testing.T) {
		testUpdate(t)
	})
	t.Run("delete", func(t *testing.T) {
		testDelete(t)
	})
	t.Run("count", func(t *testing.T) {
		testCount(t)
	})
	t.Run("orderby_desc", func(t *testing.T) {
		testOrderByDesc(t)
	})
}
