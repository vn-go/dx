package mssql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

func (d *mssqlDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {
	if !d.isReleaseMode {
		panic(fmt.Sprintf("%s not implement at mssqlDialect.SqlFunction, see %s", delegator.FuncName, `dialect\mssql\dialect.mssql.SqlFunction.go`))
	}
	return "", nil
	// //delegator.Approved = true
	// delegator.FuncName = strings.ToUpper(delegator.FuncName)
	// return "", nil
}
