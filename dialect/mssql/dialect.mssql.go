package mssql

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/internal"
)

type MssqlDialect struct {
	cacheMakeSqlInsert sync.Map
}

func (d *MssqlDialect) LikeValue(val string) string {

	return val
}
func (d *MssqlDialect) Quote(name ...string) string {
	return "[" + strings.Join(name, "].[") + "]"
}
func (d *MssqlDialect) Name() string {
	return "mssql"
}
func (d *MssqlDialect) GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error) {
	panic(fmt.Errorf("not implemented, see file eorm/dialect.mssql.go"))
}
func (d *MssqlDialect) ToText(value string) string {
	return fmt.Sprintf("N'%s'", value)
}
func (d *MssqlDialect) ToParam(index int) string {
	return fmt.Sprintf("@p%d", index)
}
func (d *MssqlDialect) SqlFunction(delegator *internal.DialectDelegateFunction) (string, error) {
	//delegator.Approved = true
	delegator.FuncName = strings.ToUpper(delegator.FuncName)
	return "", nil
}
