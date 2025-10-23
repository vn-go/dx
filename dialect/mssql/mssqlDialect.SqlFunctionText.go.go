package mssql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

// mssqlDialect.SqlFunctionText.go
func (mssql *mssqlDialect) SqlFunctionText(delegator *types.DialectDelegateFunction) (string, error) {
	if len(delegator.Args) == 1 {
		return mssql.SqlFunctionTextOneArg(delegator)
	} else if len(delegator.Args) == 2 {
		//  format
		/*
						SELECT FORMAT(Price, '000.00') AS FormattedPrice
			FROM Items;
		*/
		field := delegator.Args[0]
		formatStr := delegator.Args[1]
		delegator.HandledByDialect = true
		return fmt.Sprintf("FORMAT(%s, %s)", field, formatStr), nil
	} else {
		return "", types.NewCompilerError("You can use the TEXT function with one or two inputs.")
	}
}

func (mssql *mssqlDialect) SqlFunctionTextOneArg(delegator *types.DialectDelegateFunction) (string, error) {
	delegator.HandledByDialect = true
	return fmt.Sprintf("cast(%s as nvarchar(max))", delegator.Args[0]), nil
}
