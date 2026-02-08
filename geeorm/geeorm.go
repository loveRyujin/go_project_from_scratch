package geeorm

import (
	"database/sql"
	"log/slog"

	"github.com/loveRyujin/geeorm/dialect"
	"github.com/loveRyujin/geeorm/session"
)

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
		return nil, err
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
