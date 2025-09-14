package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/types"
)

type postgresDialect struct {
	cacheMakeSqlInsert sync.Map
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
func (d *postgresDialect) ToParam(index int) string {
	return fmt.Sprintf("$%d", index)
}
func (d *postgresDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {

	switch strings.ToLower(delegator.FuncName) {
	case "len":
		delegator.FuncName = "LENGTH"
		delegator.HandledByDialect = true
		return "LENGTH" + "(" + strings.Join(delegator.Args, ", ") + ")", nil
	case "concat":
		delegator.HandledByDialect = true
		castArgs := make([]string, len(delegator.Args))
		for i, x := range delegator.Args {
			if x[0] == '$' {
				castArgs[i] = x + "::text"
			} else {
				castArgs[i] = x
			}
		}
		return "CONCAT" + "(" + strings.Join(castArgs, ", ") + ")", nil
	default:

		return "", nil
	}
}

var postgresDialectInstance = &postgresDialect{
	cacheMakeSqlInsert: sync.Map{},
}

func NewPostgresDialect() types.Dialect {

	return postgresDialectInstance
}
