package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/internal"
)

type PostgresDialect struct {
	cacheMakeSqlInsert sync.Map
}

func (d *PostgresDialect) LikeValue(val string) string {

	return internal.ReplaceStarWithCache("postgres", val, '*', '%')
}
func (d *PostgresDialect) Name() string {
	return "postgres"
}
func (d *PostgresDialect) Quote(name ...string) string {
	return "\"" + strings.Join(name, "\".\"") + "\""
}
func (d *PostgresDialect) GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error) {
	panic(fmt.Errorf("not implemented, see file eorm/dialect.mssql.go"))
}
func (d *PostgresDialect) ToText(value string) string {
	return fmt.Sprintf("'%s'::citext", value)
}
func (d *PostgresDialect) ToParam(index int) string {
	return fmt.Sprintf("$%d", index)
}
func (d *PostgresDialect) SqlFunction(delegator *internal.DialectDelegateFunction) (string, error) {
	switch delegator.FuncName {
	case "LEN":
		delegator.FuncName = "LENGTH"
		delegator.HandledByDialect = true
		return "LENGTH" + "(" + strings.Join(delegator.Args, ", ") + ")", nil

	default:

		return "", nil
	}
}
