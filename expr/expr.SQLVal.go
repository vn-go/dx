package expr

import (
	"strconv"

	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) SQLVal(context *exprCompileContext, expr *sqlparser.SQLVal) (string, error) {
	switch expr.Type {
	case sqlparser.StrVal:
		return context.Dialect.ToText(string(expr.Val)), nil
	case sqlparser.IntVal:
		return string(expr.Val), nil
	case sqlparser.FloatVal:
		return string(expr.Val), nil
	case sqlparser.ValArg:
		if context.paramIndex == 0 {
			context.paramIndex = 1
		}

		strIndex := string(expr.Val[2:len(expr.Val)])
		if _, err := strconv.Atoi(strIndex); err == nil {
			defer func() {
				context.paramIndex++
			}()
			return context.Dialect.ToParam(context.paramIndex), nil
		} else {
			return string(expr.Val), nil
		}

	}

	return string(expr.Val), nil

}
