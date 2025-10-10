package compiler

import (
	"strconv"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) sqlVal(expr *sqlparser.SQLVal, cmpType COMPILER, args *internal.SqlArgs) (string, error) {
	switch expr.Type {
	case sqlparser.StrVal:
		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_CONSTANT,
			Value:     internal.Helper.TrimStringLiteral(string(expr.Val)), //cmp.dialect.ToText(string(expr.Val)),
			Index:     len(*args),
		})
		return "?", nil
	case sqlparser.IntVal:
		if cmpType == C_LIMIT || cmpType == C_OFFSET {
			return string(expr.Val), nil
		}
		intVal, err := internal.Helper.ToIntFormBytes(expr.Val)
		if err != nil {
			return "", err
		}

		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_CONSTANT,
			Value:     intVal,
			Index:     len(*args),
		})
		return "?", nil
	case sqlparser.FloatVal:
		floatVal, err := internal.Helper.ToFloatFormBytes(expr.Val)
		if err != nil {
			return "", err
		}
		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_CONSTANT,
			Value:     floatVal,
			Index:     len(*args),
		})
		return "?", nil

	case sqlparser.ValArg:
		strIndex := string(expr.Val[2:])
		index, err := strconv.Atoi(strIndex)
		if err != nil {
			return "", NewCompilerError("Expression is invalid")
		}
		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_DEFAULT,
			Index:     index - 1,
		})
		return "?", nil

	}

	return string(expr.Val), nil
}
