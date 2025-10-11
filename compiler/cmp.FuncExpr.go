package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) funcExpr(expr *sqlparser.FuncExpr, cmpType COMPILER, args *internal.SqlArgs) (string, error) {
	strArgs := []string{}
	if expr.Name.String() == internal.FnMarkSpecialTextArgs && len(expr.Exprs) == 1 {
		n := expr.Exprs[0].(*sqlparser.AliasedExpr).Expr.(*sqlparser.SQLVal)
		index, err := internal.Helper.ToIntFormBytes(n.Val)
		if err != nil {
			return "", fmt.Errorf("%s is not int value", string(n.Val))
		}
		indexInSql := len(*args) + 1
		*args = append(*args, internal.SqlArg{
			ParamType:   internal.PARAM_TYPE_2APOSTROPHE,
			IndexInSql:  indexInSql,
			IndexInArgs: index,
		})
		return cmp.dialect.ToParam(indexInSql), nil
	}
	for _, arg := range expr.Exprs {
		strArg, err := cmp.resolve(arg, C_FUNC, args)
		if err != nil {
			if cErr, ok := err.(*CompilerError); ok {
				if cErr.errType == ERR_TABLE_NOT_FOUND {
					return "", fmt.Errorf("can not determine table of %s in sql \n%s", arg, cmp.sql)
				}

			}
			return "", err
		} else {
			strArgs = append(strArgs, strArg)
		}

	}
	dialectDelegateFunction := types.DialectDelegateFunction{
		FuncName:         expr.Name.String(),
		Args:             strArgs,
		HandledByDialect: false,
	}
	ret, err := cmp.dialect.SqlFunction(&dialectDelegateFunction)
	if err != nil {

		return "", err
	}
	if dialectDelegateFunction.HandledByDialect {

		return ret, nil
	}
	return dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")", nil

}
