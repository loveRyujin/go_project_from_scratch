package clause

import "strings"

// Type 表示 SQL 子句的类型。
type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
	UPDATE
	DELETE
	COUNT
)

// Clause 用于分步构建 SQL 语句，各子句独立设置，最后按指定顺序拼接。
type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]any
}

// Set 设置指定类型的 SQL 子句。
//
//	c.Set(clause.SELECT, "User", []string{"*"})
//	c.Set(clause.WHERE, "Name = ?", "Tom")
func (c *Clause) Set(name Type, vars ...any) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]any)
	}
	sql, sqlVars := generators[name](vars...)
	c.sql[name] = sql
	c.sqlVars[name] = sqlVars
}

// Build 按指定顺序拼接已设置的子句，返回完整 SQL 和参数。
//
//	sql, vars := c.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
//	// "SELECT * FROM User WHERE Name = ? ORDER BY Age ASC LIMIT ?" ["Tom", 3]
func (c *Clause) Build(orders ...Type) (string, []any) {
	var sqls []string
	var vars []any
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
