package compiler

import "github.com/vn-go/dx/sqlparser"

func (cmp *compiler) aliasedTableExpr(expr *sqlparser.AliasedTableExpr, cmpType COMPILER, args *[]any) (string, error) {
	ret, err := cmp.resolve(expr.Expr, cmpType, args)
	if err != nil {
		return "", err
	}
	if _, ok := expr.Expr.(*sqlparser.Subquery); ok {
		return "(" + ret + ") " + cmp.dialect.Quote(expr.As.String()), nil
	}
	return ret, nil

}
