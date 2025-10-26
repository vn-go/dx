package mysql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

// mySqlDialect.SqlFunction.List.go
func (d *mySqlDialect) SqlFunctionResolveArrayFunctions(funcName string, delegator *types.DialectDelegateFunction) (string, error) {
	//JSON_OVERLAPS(department.children_id, JSON_ARRAY(2, 4))
	switch funcName {
	case "contains":
		delegator.HandledByDialect = true
		return fmt.Sprintf("JSON_OVERLAPS(%s, CAST(%s AS JSON))", delegator.Args[0], delegator.Args[1]), nil
	default:
		panic(fmt.Sprintf("not implement %s. ref mySqlDialect.SqlFunctionResolveArrayFunctions", funcName))
	}

}
