package mysql

import (
	"strings"

	"github.com/vn-go/dx/internal"
)

func (d *MysqlDialect) SqlFunction(delegator *internal.DialectDelegateFunction) (string, error) {
	fnName := strings.ToLower(delegator.FuncName)
	switch fnName {
	case "now":
		delegator.HandledByDialect = true
		return "NOW()", nil
	case "len":
		delegator.FuncName = "LENGTH"
		return delegator.FuncName, nil
	default:
		return "", nil
	}

}
