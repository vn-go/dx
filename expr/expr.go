package expr

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

type exprReceiver struct {
}

func (e *exprReceiver) compile(context *exprCompileContext, expr interface{}) (string, error) {
	switch expr := expr.(type) {
	case *sqlparser.ComparisonExpr:
		return e.ComparisonExpr(context, *expr)
	case *sqlparser.ColName:
		return e.ColName(context, *expr)
	case *sqlparser.AndExpr:
		return e.AndExpr(context, expr)
	case *sqlparser.OrExpr:
		return e.OrExpr(context, expr)
	case *sqlparser.SQLVal:
		return e.SQLVal(context, expr)
	case *sqlparser.StarExpr:
		return e.StarExpr(context, expr)

	case *sqlparser.JoinTableExpr:
		return e.JoinTableExpr(context, expr)
	case sqlparser.JoinTableExpr:
		return e.JoinTableExpr(context, &expr)
	case *sqlparser.AliasedTableExpr:
		return e.AliasedTableExpr(context, expr)
	case sqlparser.JoinCondition:
		return e.JoinCondition(context, &expr)
	case sqlparser.TableName:
		return e.TableName(context, &expr)
	case *sqlparser.BinaryExpr:
		return e.BinaryExpr(context, expr)
	case *sqlparser.FuncExpr:
		return e.FuncExpr(context, expr)
	case *sqlparser.AliasedExpr:
		return e.AliasedExpr(context, expr)
	case *sqlparser.Order:
		return e.Order(context, expr)
	case sqlparser.Order:
		return e.Order(context, &expr)
	case sqlparser.OrderBy:
		return e.OrderBy(context, expr)
	case *sqlparser.UpdateExpr:
		return e.UpdateExpr(context, expr)

	default:
		panic(fmt.Errorf("unsupported expression type %T in file eorm/expr.go, line 17", expr))
	}
	return "", nil
}

var exprs = &exprReceiver{}
