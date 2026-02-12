package session

import "log/slog"

// Begin 开启数据库事务。开启后 Session 的所有操作都在事务内执行。
func (s *Session) Begin() error {
	slog.Info("transcation begin")
	tx, err := s.raw.Begin()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	s.tx = tx
	return nil
}

// Commit 提交事务。提交后 Session 恢复为非事务模式。
func (s *Session) Commit() error {
	slog.Info("transcation commit")
	if err := s.tx.Commit(); err != nil {
		slog.Error(err.Error())
		return err
	}
	s.tx = nil
	return nil
}

// Rollback 回滚事务。回滚后 Session 恢复为非事务模式。
func (s *Session) Rollback() error {
	slog.Info("transcation rollback")
	if err := s.tx.Rollback(); err != nil {
		slog.Error(err.Error())
		return err
	}
	s.tx = nil
	return nil
}
