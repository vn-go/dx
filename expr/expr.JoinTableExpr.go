package expr

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) SimpleTableExpr(context *exprCompileContext, expr sqlparser.SimpleTableExpr) (string, error) {

	panic(fmt.Sprintf("not implement, exprReceiver.SimpleTableExpr %s", `expr\expr.JoinTableExpr.go`))
}
func (compiler *exprReceiver) JoinTableExpr(context *exprCompileContext, expr *sqlparser.JoinTableExpr) (string, error) {
	var left, right, on string
	var err error

	left, err = compiler.compile(context, expr.LeftExpr)
	if err != nil {
		return "", err
	}
	right, err = compiler.compile(context, expr.RightExpr)
	if err != nil {
		return "", err
	}
	on, err = compiler.compile(context, expr.Condition)
	if err != nil {
		return "", err
	}
	if expr.Join == "join" {
		if left == "" {
			return fmt.Sprintf("INNER JOIN %s ON %s", right, on), nil
		}
		return fmt.Sprintf("%s INNER JOIN %s ON %s", left, right, on), nil
	}
	if expr.Join == "left join" {
		if left == "" {
			return fmt.Sprintf("LEFT INNER JOIN %s ON %s", right, on), nil
		}
		return fmt.Sprintf("%s LEFT JOIN %s ON %s", left, right, on), nil
	}
	panic(fmt.Errorf("unsupported join type is %s, file %s, line %d", expr.Join, "eorm/expr.JoinTableExpr.go", 40))

}
