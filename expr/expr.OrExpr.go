package expr

import "github.com/vn-go/dx/sqlparser"

func (compiler *exprReceiver) OrExpr(context *exprCompileContext, expr *sqlparser.OrExpr) (string, error) {
	left, err := compiler.compile(context, expr.Left)
	if err != nil {
		return "", err

	}
	right, err := compiler.compile(context, expr.Right)
	if err != nil {
		return "", err
	}
	return left + " OR " + right, nil

}
