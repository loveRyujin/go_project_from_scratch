package geeorm

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/loveRyujin/geeorm/dialect"
	"github.com/loveRyujin/geeorm/session"
)

// TxFn 是事务回调函数类型。在 Transcation 中执行，返回 error 时自动回滚。
type TxFn func(*session.Session) (any, error)

// Engine 是 geeorm 的核心入口，管理数据库连接和 dialect。
type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

// NewEngine 创建一个新的 Engine 实例，建立数据库连接。
//
//	engine, err := geeorm.NewEngine("sqlite3", "gee.db")
//	defer engine.Close()
func NewEngine(driver, source string) (*Engine, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	if err = db.Ping(); err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	d, ok := dialect.GetDialect(driver)
	if !ok {
		slog.Error("dialect not found", slog.String("driver", driver))
		return nil, fmt.Errorf("dialect %q not found", driver)
	}

	slog.Info("Create database connection successfully")
	return &Engine{db: db, dialect: d}, nil
}

// Close 关闭数据库连接。
func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("Close database connection successfully")
}

// Session 创建一个新的 Session，用于执行数据库操作。
//
//	s := engine.Session()
//	s.Model(&User{}).CreateTable()
func (e *Engine) Session() *session.Session {
	return session.New(e.db, e.dialect)
}

// Transcation 在事务中执行 fn。fn 返回 error 或 panic 时自动回滚，否则自动提交。
//
//	result, err := engine.Transcation(func(s *session.Session) (any, error) {
//	    s.Model(&User{})
//	    _, err := s.Insert(&User{"Tom", 18})
//	    return nil, err
//	})
func (e *Engine) Transcation(fn TxFn) (result any, err error) {
	s := e.Session()
	if bErr := s.Begin(); bErr != nil {
		slog.Error("transaction begin failed", slog.String("err", bErr.Error()))
		return nil, bErr
	}
	defer func() {
		if r := recover(); r != nil {
			_ = s.Rollback()
			panic(r)
		} else if err != nil {
			_ = s.Rollback()
		} else if err = s.Commit(); err != nil {
			_ = s.Rollback()
		}
	}()

	return fn(s)
}

func (e *Engine) Migrate(val any) error {
	_, err := e.Transcation(func(s *session.Session) (any, error) {
		if !s.Model(val).HasTable() {
			slog.Info("table not exists, create new", slog.String("table", s.RefTable().Name))
			return nil, s.CreateTable()
		}

		table := s.RefTable()
		rows, err := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		if err != nil {
			return nil, err
		}
		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		addCols := diff(table.FieldNames, columns)
		delCols := diff(columns, table.FieldNames)
		slog.Info("migrate diff", slog.Any("addCols", addCols), slog.Any("delCols", delCols))

		for _, addCol := range addCols {
			f := table.GetField(addCol)
			sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table.Name, f.Name, f.Type)
			if _, err := s.Raw(sql).Exec(); err != nil {
				return nil, err
			}
		}

		if len(delCols) == 0 {
			return nil, nil
		}
		tmpTable := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmpTable, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmpTable, table.Name))
		_, err = s.Exec()
		return nil, err
	})
	return err
}

// diff returns a - b
func diff(a, b []string) []string {
	mapB := make(map[string]struct{})
	for i := range b {
		mapB[b[i]] = struct{}{}
	}

	var diff []string
	for i := range a {
		if _, ok := mapB[a[i]]; !ok {
			diff = append(diff, a[i])
		}
	}

	return diff
}
