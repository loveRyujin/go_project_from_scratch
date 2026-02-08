package schemas

import (
	"fmt"
	"go/ast"
	"reflect"

	"github.com/loveRyujin/geeorm/dialect"
)

// Field represents a column of table
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema represents a table of database
type Schema struct {
	Model      any
	Name       string
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

// Parse 将结构体解析为 Schema。dest 必须是结构体或结构体指针。
// 跳过匿名字段和未导出字段，通过 `geeorm` tag 设置约束（如 PRIMARY KEY）。
//
//	schema, err := Parse(&User{}, dialect)
func Parse(dest any, d dialect.Dialect) (*Schema, error) {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("geeorm: Parse requires a struct, got %s", modelType.Kind())
	}
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		f := modelType.Field(i)
		if !f.Anonymous && ast.IsExported(f.Name) {
			field := &Field{
				Name: f.Name,
				Type: d.DataTypeof(reflect.Indirect(reflect.New(f.Type))),
			}
			if v, ok := f.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, f.Name)
			schema.fieldMap[f.Name] = field
		}
	}
	return schema, nil
}

// GetField 根据字段名获取 Field，不存在时返回 nil。
func (s *Schema) GetField(name string) *Field {
	return s.fieldMap[name]
}

// RecordValues 提取结构体中各字段的值，按 FieldNames 顺序返回。
// val 必须是结构体或结构体指针。
func (s *Schema) RecordValues(val any) ([]any, error) {
	destValue := reflect.Indirect(reflect.ValueOf(val))
	if destValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("geeorm: RecordValues requires a struct, got %s", destValue.Kind())
	}
	fieldValue := make([]any, 0, len(s.Fields))
	for _, fname := range s.FieldNames {
		fieldValue = append(fieldValue, destValue.FieldByName(fname).Interface())
	}
	return fieldValue, nil
}
