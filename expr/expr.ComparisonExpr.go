package expr

import "github.com/vn-go/dx/sqlparser"

//ComparisonExpr
func (e *exprReceiver) ComparisonExpr(context *exprCompileContext, expr sqlparser.ComparisonExpr) (string, error) {
	left, err := e.compile(context, expr.Left)
	if err != nil {
		return "", err
	}
	right, err := e.compile(context, expr.Right)
	if err != nil {
		return "", err
	}
	ret := left + " " + expr.Operator + " " + right
	return ret, nil

}
