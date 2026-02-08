package session

import (
	"database/sql"
	"log/slog"
	"strings"

	"github.com/loveRyujin/geeorm/clause"
	"github.com/loveRyujin/geeorm/dialect"
	schemas "github.com/loveRyujin/geeorm/schema"
)

// Session 封装了数据库操作的核心会话，负责 SQL 构建与执行。
type Session struct {
	raw      *sql.DB
	clause   clause.Clause
	sql      strings.Builder
	sqlVars  []any
	dialect  dialect.Dialect
	refTable *schemas.Schema
}

// New 创建一个新的 Session 实例。一般通过 Engine.Session() 获取，无需直接调用。
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{raw: db, dialect: dialect}
}

// Clear 重置 SQL 语句和参数，每次 Exec/QueryRow/QueryRows 执行后自动调用。
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

// DB 返回底层的 *sql.DB 实例。
func (s *Session) DB() *sql.DB {
	return s.raw
}

// Raw 设置原始 SQL 语句和参数，支持链式调用。
//
//	s.Raw("SELECT * FROM User WHERE Name = ?", "Tom").QueryRow()
func (s *Session) Raw(sql string, values ...any) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec 执行 SQL 语句（INSERT/UPDATE/DELETE 等），返回 sql.Result。
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()

	slog.Info(s.sql.String(), slog.Any("args", s.sqlVars))
	result, err = s.raw.Exec(s.sql.String(), s.sqlVars...)
	if err != nil {
		slog.Error(err.Error())
	}
	return
}

// QueryRow 执行查询并返回单行结果。
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()

	slog.Info(s.sql.String(), slog.Any("args", s.sqlVars))
	return s.raw.QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 执行查询并返回多行结果。
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()

	slog.Info(s.sql.String(), slog.Any("args", s.sqlVars))
	rows, err = s.raw.Query(s.sql.String(), s.sqlVars...)
	if err != nil {
		slog.Error(err.Error())
	}
	return
}
