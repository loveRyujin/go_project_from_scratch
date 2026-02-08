package session

import (
	"fmt"
	"reflect"

	"github.com/loveRyujin/geeorm/clause"
)

func (s *Session) Insert(vals ...any) (int64, error) {
	recordValues := make([]any, 0)
	for _, val := range vals {
		table := s.Model(val).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		values, err := table.RecordValues(val)
		if err != nil {
			return 0, err
		}
		recordValues = append(recordValues, values)
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, sqlVars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, sqlVars...).Exec()
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (s *Session) Find(val any) error {
	dest := reflect.ValueOf(val)
	if dest.Kind() != reflect.Pointer || dest.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("geeorm: Find requires a pointer to slice, got %T", val)
	}
	destSlice := dest.Elem()

	itemType := destSlice.Type().Elem()
	if itemType.Kind() != reflect.Struct {
		return fmt.Errorf("geeorm: Find requires a slice of structs, got slice of %s", itemType.Kind())
	}

	table := s.Model(reflect.New(itemType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, sqlVars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, sqlVars...).QueryRows()
	if err != nil {
		return err
	}

	for rows.Next() {
		dest := reflect.New(itemType).Elem()
		values := make([]any, 0, len(table.FieldNames))
		for _, fname := range table.FieldNames {
			values = append(values, dest.FieldByName(fname).Addr().Interface())
		}
		if err := rows.Scan(values...); err != nil {
			return err
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}

	return rows.Close()
}
