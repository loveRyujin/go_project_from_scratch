package session

import (
	"database/sql"
	"log/slog"
	"strings"

	"github.com/loveRyujin/geeorm/dialect"
	schemas "github.com/loveRyujin/geeorm/schema"
)

type Session struct {
	raw      *sql.DB
	sql      strings.Builder
	sqlVars  []any
	dialect  dialect.Dialect
	refTable *schemas.Schema
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{raw: db, dialect: dialect}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

func (s *Session) DB() *sql.DB {
	return s.raw
}

func (s *Session) Raw(sql string, values ...any) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()

	slog.Info(s.sql.String(), slog.Any("args", s.sqlVars))
	result, err = s.raw.Exec(s.sql.String(), s.sqlVars...)
	if err != nil {
		slog.Error(err.Error())
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()

	slog.Info(s.sql.String(), slog.Any("args", s.sqlVars))
	return s.raw.QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()

	slog.Info(s.sql.String(), slog.Any("args", s.sqlVars))
	rows, err = s.raw.Query(s.sql.String(), s.sqlVars...)
	if err != nil {
		slog.Error(err.Error())
	}
	return
}
