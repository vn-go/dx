package dx

import (
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/sql"
)

func (db *DB) Smart(query string, args ...any) (*sql.SmartSqlParser, error) {
	dialect := factory.DialectFactory.Create(db.DriverName)

	return sql.Compiler.Resolve(dialect, query, args...)
}
func (db *DB) Compact(query string) (string, error) {
	ret, _, _, err := sql.Compact(query)
	return ret, err
}
func (db *DB) ParseDsl(query string, args ...any) (*sql.SmartSqlParser, error) {
	dialect := factory.DialectFactory.Create(db.DriverName)

	return sql.Compiler.Resolve(dialect, query, args...)
}

// func (db *DB) compileDsl(skip, take int, query string, args ...any) (*sql.SmartSqlParser, error) {
// 	dialect := factory.DialectFactory.Create(db.DriverName)

// 	return sql.Compiler.Resolve(dialect, query, args...)
// }
