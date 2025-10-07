package compiler

import (
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) aliasedTableExpr(expr *sqlparser.AliasedTableExpr, cmpType COMPILER, args *internal.SqlArgs) (string, error) {

	if subQr, ok := expr.Expr.(*sqlparser.Subquery); ok {
		cmpSubQuery, err := newCompilerFromSqlNode(subQr.Select, cmp.dialect)
		if err != nil {
			return "", nil
		}
		ret, err := cmpSubQuery.resolve(subQr.Select, C_SELECT, args)
		if err != nil {
			return "", nil
		}
		return "(" + ret + ") " + cmp.dialect.Quote(expr.As.String()), nil
	} else {
		ret, err := cmp.resolve(expr.Expr, cmpType, args)
		if err != nil {
			return "", err
		}
		return ret, nil
	}

}
