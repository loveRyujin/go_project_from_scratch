package session

import (
	"fmt"
	"reflect"
	"strings"

	schemas "github.com/loveRyujin/geeorm/schema"
)

// Model 绑定模型，解析结构体为表的 Schema。若类型未变则复用缓存。
// value 必须是结构体或结构体指针，否则 panic。
//
//	s.Model(&User{})
func (s *Session) Model(value any) *Session {
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		var err error
		s.refTable, err = schemas.Parse(value, s.dialect)
		if err != nil {
			panic(err)
		}
	}
	return s
}

// RefTable 返回当前绑定的表 Schema，需先调用 Model()。
func (s *Session) RefTable() *schemas.Schema {
	return s.refTable
}

// CreateTable 根据绑定的模型创建数据表。
//
//	s.Model(&User{}).CreateTable()
func (s *Session) CreateTable() error {
	table := s.RefTable()

	columns := make([]string, 0, len(table.FieldNames))
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}

	desc := strings.Join(columns, ",")
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

// DropTable 删除绑定模型对应的数据表（IF EXISTS）。
//
//	s.Model(&User{}).DropTable()
func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

// HasTable 判断绑定模型对应的数据表是否存在。
//
//	s.Model(&User{}).HasTable() // true or false
func (s *Session) HasTable() bool {
	sql, args := s.dialect.TableExistSQL(s.RefTable().Name)
	row := s.Raw(sql, args...).QueryRow()

	var tableName string
	_ = row.Scan(&tableName)
	return tableName == s.RefTable().Name
}
