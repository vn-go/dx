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
}
type initReplaceStar struct {
	once sync.Once
	val  string
}

var replaceStarCache sync.Map

func replaceStarWithCache(driver string, raw string, matche byte, replace byte) string {
	key := fmt.Sprintf("%s_%s_%d_%d", driver, raw, matche, replace)
	actual, _ := replaceStarCache.LoadOrStore(key, &initReplaceStar{})
	init := actual.(*initReplaceStar)
	init.once.Do(func() {
		init.val = replaceStar(driver, raw, matche, replace)
	})
	return init.val

}
func replaceStar(driver string, raw string, matche byte, replace byte) string {
	var builder strings.Builder
	n := len(raw)
	for i := 0; i < n; i++ {
		if raw[i] == matche {
			if i == 0 || raw[i-1] != '\\' {
				builder.WriteByte(replace)
			} else {
				builder.WriteByte(matche)
			}
		} else {
			builder.WriteByte(raw[i])
		}
	}
	return builder.String()
}

func (d *mySqlDialect) LikeValue(val string) string {
	return replaceStarWithCache("mysql", val, '*', '%')
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
	return "?"
}

var mySqlDialectInstance = &mySqlDialect{
	cacheMakeSqlInsert: sync.Map{},
}

func NewMysqlDialect() types.Dialect {
	return mySqlDialectInstance
}
