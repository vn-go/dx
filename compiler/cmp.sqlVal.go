package compiler

import (
	"strconv"

	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) sqlVal(expr *sqlparser.SQLVal, cmpType COMPILER) (string, error) {
	switch expr.Type {
	case sqlparser.StrVal:
		return cmp.dialect.ToText(string(expr.Val)), nil
	case sqlparser.IntVal:
		return string(expr.Val), nil
	case sqlparser.FloatVal:
		return string(expr.Val), nil
	case sqlparser.ValArg:
		if cmp.paramIndex == 0 {
			cmp.paramIndex = 1
		}

		strIndex := string(expr.Val[2:len(expr.Val)])
		if _, err := strconv.Atoi(strIndex); err == nil {
			defer func() {
				cmp.paramIndex++
			}()
			return cmp.dialect.ToParam(cmp.paramIndex), nil
		} else {
			return string(expr.Val), nil
		}

	}

	return string(expr.Val), nil
}
