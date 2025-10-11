package compiler

import (
	"strconv"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) sqlVal(expr *sqlparser.SQLVal, cmpType COMPILER, args *internal.SqlArgs) (string, error) {
	switch expr.Type {
	case sqlparser.StrVal:
		indexInSql := len(*args) + 1
		*args = append(*args, internal.SqlArg{
			ParamType:  internal.PARAM_TYPE_CONSTANT,
			Value:      internal.Helper.TrimStringLiteral(string(expr.Val)), //cmp.dialect.ToText(string(expr.Val)),
			IndexInSql: indexInSql,
		})
		return cmp.dialect.ToParam(indexInSql), nil
	case sqlparser.IntVal:
		if cmpType == C_LIMIT || cmpType == C_OFFSET {
			return string(expr.Val), nil
		}
		intVal, err := internal.Helper.ToIntFormBytes(expr.Val)
		if err != nil {
			return "", err
		}
		indexInSql := len(*args) + 1
		*args = append(*args, internal.SqlArg{
			ParamType:  internal.PARAM_TYPE_CONSTANT,
			Value:      intVal,
			IndexInSql: indexInSql,
		})
		return cmp.dialect.ToParam(indexInSql), nil
	case sqlparser.FloatVal:
		floatVal, err := internal.Helper.ToFloatFormBytes(expr.Val)
		if err != nil {
			return "", err
		}
		indexInSql := len(*args) + 1
		*args = append(*args, internal.SqlArg{
			ParamType:  internal.PARAM_TYPE_CONSTANT,
			Value:      floatVal,
			IndexInSql: indexInSql,
		})
		return cmp.dialect.ToParam(indexInSql), nil

	case sqlparser.ValArg:
		strIndex := string(expr.Val[2:])
		index, err := strconv.Atoi(strIndex)
		if err != nil {
			return "", NewCompilerError("Expression is invalid")
		}
		indexInSql := len(*args) + 1
		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_DEFAULT,

			IndexInSql:  indexInSql,
			IndexInArgs: index - 1,
		})
		return cmp.dialect.ToParam(index), nil

	}

	return string(expr.Val), nil
}
