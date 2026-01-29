package schemas

import (
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

func Parse(dest any, d dialect.Dialect) *Schema {
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
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
	return schema
}

func (s *Schema) GetField(name string) *Field {
	return s.fieldMap[name]
}
