package mysql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

var castFunc map[string]string = map[string]string{
	"day":         "DAY",
	"month":       "MONTH",
	"year":        "YEAR",
	"hour":        "HOUR",
	"minutes":     "MINUTE",
	"second":      "SECOND",
	"microsecond": "MICROSECOND",
	"len":         "LENGTH",
	"isnull":      "IFNULL",
	"date":        "DATE",
	"concat":      "CONCAT",
}

func (d *mySqlDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {
	fnName := strings.ToLower(delegator.FuncName)
	if ret, ok := castFunc[fnName]; ok {
		delegator.FuncName = ret
		return ret, nil
	}
	switch fnName {
	case "now":
		delegator.HandledByDialect = true
		return "NOW()", nil

	default:
		if !d.isReleaseMode {
			defer func() {
				panic(fmt.Sprintf("%s not implement at mySqlDialect.SqlFunction, see %s", delegator.FuncName, `dialect\mysql\dialect.mysql.SqlFunction.go`))
			}()

		}
		return "", nil
	}

}
