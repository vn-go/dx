package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/types"
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
func (d *mySqlDialect) ToParam(index int) string {
	return fmt.Sprintf("{%d}", index)
	//return "?"
}

var mySqlDialectInstance = &mySqlDialect{
	cacheMakeSqlInsert: sync.Map{},
}

func NewMysqlDialect() types.Dialect {

	return mySqlDialectInstance
}
