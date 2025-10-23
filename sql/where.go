package sql

import (
	"fmt"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type WhereType struct {
}

func (w *WhereType) splitAndExpr(expr sqlparser.Expr) []sqlparser.SQLNode {
	if and, ok := expr.(*sqlparser.AndExpr); ok {
		l := w.splitAndExpr(and.Left)
		r := w.splitAndExpr(and.Right)
		return append(l, r...)
	}
	if p, ok := expr.(*sqlparser.ParenExpr); ok {
		return w.splitAndExpr(p.Expr)
	}
	return []sqlparser.SQLNode{expr}
}

var where = &WhereType{}

func (w *WhereType) resolve(expr sqlparser.Expr, injector *injector, selectedExprsReverse dictionaryFields) (*compilerResult, error) {
	switch x := expr.(type) {
	case *sqlparser.ComparisonExpr:
		return w.comparisonExpr(x, injector, selectedExprsReverse)
	case *sqlparser.ColName:
		colName := x.Name.Lowered()
		if fx, ok := selectedExprsReverse[colName]; ok {
			return &compilerResult{
				Content:           fx.Expr,
				IsInAggregateFunc: fx.IsInAggregateFunc,
			}, nil
		}
		return selector.colName(x, injector)
	case *sqlparser.FuncExpr:
		if x.Name.String() == GET_PARAMS_FUNC {
			return params.funcExpr(x, injector)
		}
		return w.funcExpr(x, injector)
	case *sqlparser.AndExpr:
		return exp.resolve(x, injector, CMP_WHERE)
	case *sqlparser.BinaryExpr:
		return exp.resolve(x, injector, CMP_WHERE)
	case *sqlparser.SQLVal:
		return params.sqlVal(x, injector)
	default:
		panic(fmt.Sprintf("unsupported expression type %T. See WhereType.Resolve", x))
	}

}

func (w *WhereType) funcExpr(x *sqlparser.FuncExpr, injector *injector) (*compilerResult, error) {
	if x.Name.String() == internal.FnMarkSpecialTextArgs {
		return params.funcExpr(x, injector)
	}

	return exp.funcExpr(x, injector, CMP_WHERE)
}
