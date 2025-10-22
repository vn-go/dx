package sql

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

type WhereType struct {
}

var where = &WhereType{}

func (w *WhereType) resolve(expr sqlparser.Expr, injector *injector) (*compilerResult, error) {
	switch x := expr.(type) {
	case *sqlparser.ComparisonExpr:
		return w.comparisonExpr(x, injector)
	case *sqlparser.ColName:
		return selector.colName(x, injector)
	case *sqlparser.FuncExpr:
		if x.Name.String() == GET_PARAMS_FUNC {
			return params.funcExpr(x, injector)
		}
		return w.funcExpr(x, injector)
	case *sqlparser.AndExpr:
		return exp.resolve(x, injector)
	default:
		panic(fmt.Sprintf("unsupported expression type %T. See WhereType.Resolve", x))
	}

}

func (w *WhereType) funcExpr(x *sqlparser.FuncExpr, injector *injector) (*compilerResult, error) {
	panic("unimplemented")
}
