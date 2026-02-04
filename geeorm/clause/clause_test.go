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
	clause.Set(ORDERBY, "Age ASC")
	sql, vars := clause.Build(SELECT, WHERE, ORDERBY, LIMIT)
	t.Log(sql, vars)
	if sql != "SELECT * FROM User WHERE Name = ? ORDER BY Age ASC LIMIT ?" {
		t.Fatal("failed to build SQL")
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

func TestClause_Build(t *testing.T) {
	t.Run("select", func(t *testing.T) {
		testSelect(t)
	})
	t.Run("insert", func(t *testing.T) {
		testInsert(t)
	})
}
