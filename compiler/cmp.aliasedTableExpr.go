package compiler

import "github.com/vn-go/dx/sqlparser"

func (cmp *compiler) aliasedTableExpr(expr *sqlparser.AliasedTableExpr, cmpType COMPILER) (string, error) {
	ret, err := cmp.resolve(expr.Expr, cmpType)
	if err != nil {
		return "", err
	}
	return ret, nil

}
