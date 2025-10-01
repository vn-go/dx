package mssql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

func (d *mssqlDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {

	if !d.isReleaseMode {
		defer func() {
			panic(fmt.Sprintf("%s not implement at mssqlDialect.SqlFunction, see %s", delegator.FuncName, `dialect\mssql\dialect.mssql.SqlFunction.go`))
		}()
		return "", fmt.Errorf("%s is not function", delegator.FuncName)
	} else {
		return "", fmt.Errorf("%s is not function", delegator.FuncName)
	}
	// //delegator.Approved = true
	// delegator.FuncName = strings.ToUpper(delegator.FuncName)
	// return "", nil
}
