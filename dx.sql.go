package dx

import (
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/sql"
)

func (db *DB) Smart(query string, args ...any) (*sql.SmartSqlParser, error) {
	dialect := factory.DialectFactory.Create(db.DriverName)

	return sql.Compiler.Resolve(dialect, query, args...)
}
