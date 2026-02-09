package session

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/loveRyujin/geeorm/clause"
)

// Insert 插入一条或多条记录，返回影响行数。
// vals 必须是结构体或结构体指针。
//
//	s.Insert(&User{"Tom", 18})                           // 单条插入
//	s.Insert(&User{"Tom", 18}, &User{"Sam", 25})         // 批量插入
func (s *Session) Insert(vals ...any) (int64, error) {
	recordValues := make([]any, 0)
	for _, val := range vals {
		table := s.Model(val).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)

		if err := s.CallMethod(BeforeInsert, val); err != nil {
			return 0, err
		}

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

	for _, val := range vals {
		if err := s.CallMethod(AfterInsert, val); err != nil {
			return 0, err
		}
	}

	return result.RowsAffected()
}

// Find 查询所有匹配的记录，结果填充到传入的切片指针中。
// val 必须是指向结构体切片的指针。
//
//	var users []User
//	s.Find(&users)
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

	if err := s.CallMethod(BeforeQuery, table.Model); err != nil {
		return err
	}

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
		if err := s.CallMethod(AfterQuery, dest.Addr().Interface()); err != nil {
			return err
		}
		destSlice.Set(reflect.Append(destSlice, dest))
	}

	return rows.Close()
}

// Update 更新记录，返回影响行数。支持三种调用方式：
//
//	s.Update(&User{Name: "Tom", Age: 18})                // struct
//	s.Update(map[string]any{"Name": "Tom", "Age": 18})   // map
//	s.Update("Name", "Tom", "Age", 18)                   // kv 平铺
//
// 通常配合 WHERE 条件使用：
//
//	s.clause.Set(clause.WHERE, "Name = ?", "Tom")
//	s.Update("Age", 25)
func (s *Session) Update(vals ...any) (int64, error) {
	if err := s.CallMethod(BeforeUpdate, s.RefTable().Model); err != nil {
		return 0, err
	}

	m, err := toUpdateMap(s.RefTable().FieldNames, vals...)
	if err != nil {
		return 0, err
	}

	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, sqlVars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, sqlVars...).Exec()
	if err != nil {
		return 0, err
	}

	if err := s.CallMethod(AfterUpdate, s.RefTable().Model); err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// toUpdateMap 将三种输入统一转换为 map[string]any
func toUpdateMap(fieldNames []string, vals ...any) (map[string]any, error) {
	if len(vals) == 0 {
		return nil, fmt.Errorf("geeorm: Update requires at least one argument")
	}

	// 只传一个参数时，支持 struct 或 map
	if len(vals) == 1 {
		switch v := vals[0].(type) {
		case map[string]any:
			return v, nil
		default:
			// 尝试当作 struct 处理
			rv := reflect.Indirect(reflect.ValueOf(vals[0]))
			if rv.Kind() != reflect.Struct {
				return nil, fmt.Errorf("geeorm: Update single arg must be a struct or map[string]any, got %T", vals[0])
			}
			m := make(map[string]any, rv.NumField())
			for _, name := range fieldNames {
				m[name] = rv.FieldByName(name).Interface()
			}
			return m, nil
		}
	}

	// 多个参数时，当作 kv 平铺处理
	if len(vals)%2 != 0 {
		return nil, fmt.Errorf("geeorm: Update kv pairs must be even, got %d args", len(vals))
	}
	m := make(map[string]any, len(vals)/2)
	for i := 0; i < len(vals); i += 2 {
		key, ok := vals[i].(string)
		if !ok {
			return nil, fmt.Errorf("geeorm: Update kv key must be string, got %T", vals[i])
		}
		m[key] = vals[i+1]
	}
	return m, nil
}

// Delete 删除记录，返回影响行数。通常配合 WHERE 条件使用。
//
//	s.clause.Set(clause.WHERE, "Name = ?", "Tom")
//	s.Delete()
func (s *Session) Delete() (int64, error) {
	if err := s.CallMethod(BeforeDelete, s.RefTable().Model); err != nil {
		return 0, err
	}

	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, sqlVars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, sqlVars...).Exec()
	if err != nil {
		return 0, err
	}

	if err := s.CallMethod(AfterDelete, s.RefTable().Model); err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Count 返回匹配的记录总数。
//
//	total, err := s.Count() // SELECT count(*) FROM User
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, sqlVars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, sqlVars...).QueryRow()

	var num int64
	if err := row.Scan(&num); err != nil {
		return 0, err
	}
	return num, nil
}

// Limit 设置查询的最大返回行数，支持链式调用。
//
//	var users []User
//	s.Limit(3).Find(&users)
func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

// Where 设置查询/更新/删除的 WHERE 条件，支持链式调用。
//
//	s.Where("Name = ?", "Tom").Find(&users)
//	s.Where("Age > ?", 18).Delete()
func (s *Session) Where(desc string, args ...any) *Session {
	query := append([]any{desc}, args...)
	s.clause.Set(clause.WHERE, query...)
	return s
}

// OrderBy 设置查询的排序规则，desc 为 true 时降序排列，支持链式调用。
//
//	s.OrderBy("Age", false).Find(&users)  // ORDER BY Age
//	s.OrderBy("Age", true).Find(&users)   // ORDER BY Age DESC
func (s *Session) OrderBy(field string, desc bool) *Session {
	s.clause.Set(clause.ORDERBY, field, desc)
	return s
}

// First 查询第一条匹配的记录，结果填充到传入的结构体指针中。
// val 必须是结构体指针。未找到记录时返回 "record not found" 错误。
//
//	var user User
//	s.Where("Name = ?", "Tom").First(&user)
func (s *Session) First(val any) error {
	dest := reflect.Indirect(reflect.ValueOf(val))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("record not found")
	}
	dest.Set(destSlice.Index(0))
	return nil
}
