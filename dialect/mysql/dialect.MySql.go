package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

type mySqlDialect struct {
	cacheMakeSqlInsert sync.Map
	isReleaseMode      bool
}

var mysqlBoolMap = map[string]string{
	"yes":   "TRUE",
	"true":  "TRUE",
	"no":    "FALSE",
	"false": "FALSE",
}

func (d *mySqlDialect) ReleaseMode(v bool) {
	d.isReleaseMode = v
}
func (d *mySqlDialect) ToBool(val string) string {
	if ret, ok := mysqlBoolMap[strings.ToLower(val)]; ok {
		return ret
	}
	return val
}
func (d *mySqlDialect) LikeValue(val string) string {
	return types.ReplaceStarWithCache("mysql", val, '*', '%')
}
func (d *mySqlDialect) Quote(name ...string) string {
	return "`" + strings.Join(name, "`.`") + "`"
}
func (d *mySqlDialect) Name() string {
	return "mysql"
}
func (d *mySqlDialect) GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error) {
	panic(fmt.Errorf("not implemented, see file eorm/dialect.mssql.go"))
}
func (d *mySqlDialect) ToText(value string) string {
	return fmt.Sprintf("'%s'", value)
}
func (d *mySqlDialect) ToParam(index int, pType sqlparser.ValType) string {
	return fmt.Sprintf("{%d}", index)
	//return "?"
}
func (d *mySqlDialect) GetSelectStatement(stmt types.SelectStatement) string {
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
	if stmt.Limit != nil {
		// MySQL: LIMIT limit OFFSET offset
		if stmt.Offset != nil {
			sql += fmt.Sprintf(" LIMIT %s OFFSET %s", stmt.Limit.Content, stmt.Offset.Content)
		} else {
			sql += fmt.Sprintf(" LIMIT %s", stmt.Limit.Content)
		}
	}

	return sql
}

var mySqlDialectInstance = &mySqlDialect{
	cacheMakeSqlInsert: sync.Map{},
}

func NewMysqlDialect() types.Dialect {

	return mySqlDialectInstance
}
