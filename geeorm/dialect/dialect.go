package dialect

import "reflect"

var dialectMap = map[string]Dialect{}

type Dialect interface {
	DataTypeof(typ reflect.Value) string
	TableExistSQL(tableName string) (string, []any)
}

func RegisterDialect(name string, dialect Dialect) {
	dialectMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectMap[name]
	return
}
