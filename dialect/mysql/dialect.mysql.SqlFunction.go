package mysql

import (
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

func (d *mySqlDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {
	fnName := strings.ToLower(delegator.FuncName)
	switch fnName {
	case "now":
		delegator.HandledByDialect = true
		return "NOW()", nil
	case "len":
		delegator.FuncName = "LENGTH"
		return delegator.FuncName, nil
	case "concat":
		newArs := []string{}
		for _, x := range delegator.Args {
			newArs = append(newArs, "IFNULL("+x+", '')")
		}
		delegator.Args = newArs
		return "", nil

	default:
		return "", nil
	}

}
