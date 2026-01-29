package session

import (
	"fmt"
	"reflect"
	"strings"

	schemas "github.com/loveRyujin/geeorm/schema"
)

func (s *Session) Model(value any) *Session {
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		s.refTable = schemas.Parse(value, s.dialect)
	}
	return s
}

func (s *Session) RefTable() *schemas.Schema {
	return s.refTable
}

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

func (s *Session) DropTable() error {
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

func (s *Session) HasTable() bool {
	sql, args := s.dialect.TableExistSQL(s.RefTable().Name)
	row := s.Raw(sql, args...).QueryRow()

	var tableName string
	_ = row.Scan(&tableName)
	return tableName == s.RefTable().Name
}
