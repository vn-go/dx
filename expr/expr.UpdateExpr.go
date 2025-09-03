package expr

import "github.com/vn-go/dx/sqlparser"

func (compiler *exprReceiver) UpdateExpr(context *exprCompileContext, expr *sqlparser.UpdateExpr) (string, error) {
	field, err := compiler.compile(context, expr.Name)
	if err != nil {
		return "", err
	}
	value, err := compiler.compile(context, expr.Expr)
	if err != nil {
		return "", err
	}
	return field + " = " + value, nil
}
