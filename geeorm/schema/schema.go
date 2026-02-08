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

func (s *Schema) GetField(name string) *Field {
	return s.fieldMap[name]
}

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
