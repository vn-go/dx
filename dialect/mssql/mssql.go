package mssql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

type mssqlDialect struct {
	cacheMakeSqlInsert sync.Map
	isReleaseMode      bool
}

var mssqlBoolMap = map[string]string{
	"yes":   "1",
	"true":  "1",
	"no":    "0",
	"false": "0",
}

func (d *mssqlDialect) ReplacePlaceholders(query string) string {
	var builder strings.Builder
	inSingle := false
	inDouble := false
	argIndex := 1

	for i := 0; i < len(query); i++ {
		ch := query[i]

		switch ch {
		case '\'':
			// Toggle trạng thái nếu không bị escape
			if !inDouble {
				inSingle = !inSingle
			}
			builder.WriteByte(ch)

		case '"':
			// Toggle trạng thái nếu không bị escape
			if !inSingle {
				inDouble = !inDouble
			}
			builder.WriteByte(ch)

		case '?':
			if inSingle || inDouble {
				// '?' nằm trong literal, giữ nguyên
				builder.WriteByte('?')
			} else {
				builder.WriteString(fmt.Sprintf("@p%d", argIndex))
				argIndex++
			}

		default:
			builder.WriteByte(ch)
		}
	}

	return builder.String()
}
func (d *mssqlDialect) ReleaseMode(v bool) {
	d.isReleaseMode = v
}
func (d *mssqlDialect) ToBool(val string) string {
	if ret, ok := mssqlBoolMap[strings.ToLower(val)]; ok {
		return ret
	}
	return val
}
func (d *mssqlDialect) LikeValue(val string) string {

	return val
}
func (d *mssqlDialect) Quote(name ...string) string {
	return "[" + strings.Join(name, "].[") + "]"
}
func (d *mssqlDialect) Name() string {
	return "mssql"
}
func (d *mssqlDialect) GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error) {
	panic(fmt.Errorf("not implemented, see file eorm/dialect.mssql.go"))
}
func (d *mssqlDialect) ToText(value string) string {
	return fmt.Sprintf("N'%s'", value)
}
func (d *mssqlDialect) ToParam(index int, pType sqlparser.ValType) string {
	return fmt.Sprintf("@p%d", index)
}


var mssqlDialectIntance = &mssqlDialect{
	cacheMakeSqlInsert: sync.Map{},
}

func NewMssqlDialect() types.Dialect {

	return mssqlDialectIntance
}
