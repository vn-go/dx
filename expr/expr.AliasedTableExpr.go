package expr

import "github.com/vn-go/dx/sqlparser"

func (compiler *exprReceiver) AliasedTableExpr(context *exprCompileContext, expr *sqlparser.AliasedTableExpr) (string, error) {
	if context.purpose == build_purpose_join {
		if expr.As.IsEmpty() {
			return compiler.compile(context, expr.Expr)
		}
		context.stackAliasTables.Push(expr.As.CompliantName())
		ret, err := compiler.compile(context, expr.Expr)
		if err != nil {
			return "", err
		}
		return ret, nil

	}
	return "", nil
	//"
	// tableName := expr.As.CompliantName()
	// if tableName == "$$$$$$$$$$$$$$" {
	// 	return "", nil
	// }
	// if tableName == "" {
	// 	return compiler.compile(context, expr.Expr)
	// } else {
	// 	if context.purpose == build_purpose_join {
	// 		context.stackAliasTables.Push(tableName)
	// 		ret, err := compiler.compile(context, expr.Expr)
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 		return ret, nil

	// 	}
	// 	return compiler.compile(context, expr.Expr)
	// }
	// return tableName, nil

}
