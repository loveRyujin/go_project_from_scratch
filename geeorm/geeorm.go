package geeorm

import (
	"database/sql"
	"log/slog"

	"github.com/loveRyujin/geeorm/session"
)

type Engine struct {
	db *sql.DB
}

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

	slog.Info("Create database connection successfully")
	return &Engine{db: db}, nil
}

func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		slog.Error(err.Error())
		return
	}
	slog.Info("Close database connection successfully")
}

func (e *Engine) Session() *session.Session {
	return session.New(e.db)
}
