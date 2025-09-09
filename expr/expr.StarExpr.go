package expr

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) StarExpr(context *exprCompileContext, expr *sqlparser.StarExpr) (string, error) {
	if expr.TableName.IsEmpty() {
		return "*", nil
	}
	panic(fmt.Sprintf("not implement, exprReceiver.StarExpr %s", `expr\expr.StarExpr.go`))

}
