package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

type postgresDialect struct {
	cacheMakeSqlInsert sync.Map
	isReleaseMode      bool
}

var pgBoolMap = map[string]string{
	"yes":   "TRUE",
	"true":  "TRUE",
	"no":    "FALSE",
	"false": "FALSE",
}

func (d *postgresDialect) ReleaseMode(v bool) {
	d.isReleaseMode = v
}

func (d *postgresDialect) ToBool(val string) string {
	if ret, ok := pgBoolMap[strings.ToLower(val)]; ok {
		return ret
	}
	return val
}
func (d *postgresDialect) LikeValue(val string) string {

	return types.ReplaceStarWithCache("postgres", val, '*', '%')
}
func (d *postgresDialect) Name() string {
	return "postgres"
}
func (d *postgresDialect) Quote(name ...string) string {
	return "\"" + strings.Join(name, "\".\"") + "\""
}
func (d *postgresDialect) GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error) {
	panic(fmt.Errorf("not implemented, see file eorm/dialect.mssql.go"))
}
func (d *postgresDialect) ToText(value string) string {
	return fmt.Sprintf("'%s'::citext", value)
}
func (d *postgresDialect) GetSelectStatement(stmt types.SelectStatement) string {
	sql := "SELECT " + stmt.Selector + " FROM " + stmt.Source

	// WHERE
	if stmt.Filter != "" {
		sql += " WHERE " + stmt.Filter
	}

	// GROUP BY
	if stmt.GroupBy != "" {
		sql += " GROUP BY " + stmt.GroupBy
	}

	// HAVING
	if stmt.Having != "" {
		sql += " HAVING " + stmt.Having
	}

	// ORDER BY
	if stmt.Sort != "" {
		sql += " ORDER BY " + stmt.Sort
	}

	// LIMIT & OFFSET
	if stmt.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", stmt.Limit)
		if stmt.Offset > 0 {
			sql += fmt.Sprintf(" OFFSET %d", stmt.Offset)
		}
	}

	return sql
}

func (d *postgresDialect) ToParam(index int, pType sqlparser.ValType) string {
	// switch pType {
	// case sqlparser.BitVal:
	// 	return fmt.Sprintf("$%d::boolean", index)
	// case sqlparser.FloatVal:
	// 	return fmt.Sprintf("$%d::float8", index)
	// case sqlparser.IntVal:
	// 	return fmt.Sprintf("$%d::bigint", index)
	// case sqlparser.StrVal:
	// 	return fmt.Sprintf("$%d::text", index)
	// default:
	// 	return fmt.Sprintf("$%d", index)
	// }
	return fmt.Sprintf("$%d", index)

}

var postgresDialectInstance = &postgresDialect{
	cacheMakeSqlInsert: sync.Map{},
}

func NewPostgresDialect() types.Dialect {

	return postgresDialectInstance
}
