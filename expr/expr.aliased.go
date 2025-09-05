package expr

import "github.com/vn-go/dx/sqlparser"

// "github.com/vn-go/dx/sqlparser"

func (compiler *exprReceiver) AliasedExpr(context *exprCompileContext, expr *sqlparser.AliasedExpr) (string, error) {
	return compiler.compile(context, expr.Expr)

}
