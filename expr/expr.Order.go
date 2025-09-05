package expr

import (
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) Order(context *exprCompileContext, expr *sqlparser.Order) (string, error) {
	return compiler.compile(context, expr.Expr)

}
func (compiler *exprReceiver) OrderBy(context *exprCompileContext, expr sqlparser.OrderBy) (string, error) {
	strOrders := []string{}
	for _, order := range expr {

		if order.Direction == "" {
			order.Direction = "ASC"

		}
		str, err := compiler.compile(context, order.Expr)
		if err != nil {
			return "", err
		}
		strOrders = append(strOrders, str+" "+strings.ToUpper(order.Direction))
	}
	return strings.Join(strOrders, ", "), nil
}
