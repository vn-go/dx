package sql

import (
	"fmt"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type expCmp struct {
}

func (e *expCmp) resolve(node sqlparser.SQLNode, injector *injector) (*compilerResult, error) {
	switch x := node.(type) {
	case *sqlparser.AndExpr:
		left, err := e.resolve(x.Left, injector)
		if err != nil {
			return nil, err
		}
		right, err := e.resolve(x.Right, injector)
		if err != nil {
			return nil, err
		}
		return &compilerResult{
			OriginalContent:      fmt.Sprintf("%s AND %s", left.Content, right.Content),
			Content:              fmt.Sprintf("%s AND %s", left.Content, right.Content),
			Args:                 append(left.Args, right.Args...),
			Fields:               left.Fields.merge(right.Fields),
			selectedExprs:        left.selectedExprs.merge(right.selectedExprs),
			selectedExprsReverse: left.selectedExprsReverse.merge(right.selectedExprsReverse),
		}, nil
	case *sqlparser.OrExpr:
		left, err := e.resolve(x.Left, injector)
		if err != nil {
			return nil, err
		}
		right, err := e.resolve(x.Right, injector)
		if err != nil {
			return nil, err
		}
		return &compilerResult{
			OriginalContent:      fmt.Sprintf("%s OR %s", left.Content, right.Content),
			Content:              fmt.Sprintf("%s OR %s", left.Content, right.Content),
			Args:                 append(left.Args, right.Args...),
			Fields:               left.Fields.merge(right.Fields),
			selectedExprs:        left.selectedExprs.merge(right.selectedExprs),
			selectedExprsReverse: left.selectedExprsReverse.merge(right.selectedExprsReverse),
		}, nil
	case *sqlparser.ComparisonExpr:
		left, err := e.resolve(x.Left, injector)
		if err != nil {
			return nil, err
		}
		right, err := e.resolve(x.Right, injector)
		if err != nil {
			return nil, err
		}
		return &compilerResult{
			OriginalContent:      fmt.Sprintf("%s %s %s", left.Content, x.Operator, right.Content),
			Content:              fmt.Sprintf("%s %s %s", left.Content, x.Operator, right.Content),
			Args:                 append(left.Args, right.Args...),
			Fields:               left.Fields.merge(right.Fields),
			selectedExprs:        left.selectedExprs.merge(right.selectedExprs),
			selectedExprsReverse: left.selectedExprsReverse.merge(right.selectedExprsReverse),
		}, nil
	case *sqlparser.ColName:
		return selector.colName(x, injector)
	case *sqlparser.FuncExpr:

		if x.Name.String() == GET_PARAMS_FUNC || x.Name.String() == internal.FnMarkSpecialTextArgs {
			return params.funcExpr(x, injector)
		} else {
			return e.FuncExpr(x, injector)
		}

	default:
		panic(fmt.Sprintf("unhandled node type %T. see  expCmp.resolve, file %s", x, `sql\where.comparisonExpr.go`))
	}

}

func (e *expCmp) FuncExpr(expr *sqlparser.FuncExpr, injector *injector) (*compilerResult, error) {
	panic(fmt.Sprintf("unhandled node type %s. see  expCmp.resolve, file %s", expr.Name.String(), `sql\where.comparisonExpr.go`))
}

var exp = &expCmp{}
