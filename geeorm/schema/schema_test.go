package schemas

import (
	"testing"

	"github.com/loveRyujin/geeorm/dialect"
)

type User struct {
	Name string `geeorm:"PRIMARY KEY"`
	Age  int
}

var testDialect, _ = dialect.GetDialect("sqlite3")

func TestParse(t *testing.T) {
	schema := Parse(&User{}, testDialect)

	if schema.Name != "User" {
		t.Errorf("expected table name 'User', got %s", schema.Name)
	}
	if len(schema.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(schema.Fields))
	}
}

func TestParse_FieldNames(t *testing.T) {
	schema := Parse(&User{}, testDialect)

	expected := []string{"Name", "Age"}
	if len(schema.FieldNames) != len(expected) {
		t.Fatalf("expected %d field names, got %d", len(expected), len(schema.FieldNames))
	}
	for i, name := range expected {
		if schema.FieldNames[i] != name {
			t.Errorf("expected field name %s at index %d, got %s", name, i, schema.FieldNames[i])
		}
	}
}

func TestParse_FieldTypes(t *testing.T) {
	schema := Parse(&User{}, testDialect)

	// Name 应该是 text 类型
	if schema.Fields[0].Type != "text" {
		t.Errorf("expected Name type 'text', got %s", schema.Fields[0].Type)
	}
	// Age 应该是 integer 类型
	if schema.Fields[1].Type != "integer" {
		t.Errorf("expected Age type 'integer', got %s", schema.Fields[1].Type)
	}
}

func TestParse_Tags(t *testing.T) {
	schema := Parse(&User{}, testDialect)

	// Name 应该有 PRIMARY KEY tag
	if schema.Fields[0].Tag != "PRIMARY KEY" {
		t.Errorf("expected Name tag 'PRIMARY KEY', got '%s'", schema.Fields[0].Tag)
	}
	// Age 没有 tag
	if schema.Fields[1].Tag != "" {
		t.Errorf("expected Age tag empty, got '%s'", schema.Fields[1].Tag)
	}
}

func TestParse_GetField(t *testing.T) {
	schema := Parse(&User{}, testDialect)

	field := schema.GetField("Name")
	if field == nil {
		t.Fatal("expected field 'Name', got nil")
	}
	if field.Name != "Name" {
		t.Errorf("expected field name 'Name', got %s", field.Name)
	}

	// 不存在的字段
	if schema.GetField("NotExist") != nil {
		t.Error("expected nil for non-existent field")
	}
}

func TestParse_PointerAndValue(t *testing.T) {
	// 传指针
	schema1 := Parse(&User{}, testDialect)
	// 传值
	schema2 := Parse(User{}, testDialect)

	if schema1.Name != schema2.Name {
		t.Error("expected same table name for pointer and value")
	}
	if len(schema1.Fields) != len(schema2.Fields) {
		t.Error("expected same fields count for pointer and value")
	}
}

type IgnoreUnexported struct {
	Name       string
	unexported int
}

func TestParse_IgnoreUnexported(t *testing.T) {
	schema := Parse(&IgnoreUnexported{}, testDialect)

	if len(schema.Fields) != 1 {
		t.Errorf("expected 1 field (unexported should be ignored), got %d", len(schema.Fields))
	}
	if schema.Fields[0].Name != "Name" {
		t.Errorf("expected field 'Name', got %s", schema.Fields[0].Name)
	}
}
