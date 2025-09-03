package dx

import (
	"strings"

	"github.com/vn-go/dx/dialect/common"
)

func (d *MysqlDialect) SqlFunction(delegator *common.DialectDelegateFunction) (string, error) {
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
