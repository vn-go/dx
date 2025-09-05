package expr

import "github.com/vn-go/dx/sqlparser"

func (compiler *exprReceiver) AndExpr(context *exprCompileContext, expr *sqlparser.AndExpr) (string, error) {
	left, err := compiler.compile(context, expr.Left)
	if err != nil {
		return "", err

	}
	right, err := compiler.compile(context, expr.Right)
	if err != nil {
		return "", err
	}
	return left + " AND " + right, nil

}
