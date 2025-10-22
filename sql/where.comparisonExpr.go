package sql

import "github.com/vn-go/dx/sqlparser"

// where.comparisonExpr.go
func (w *WhereType) comparisonExpr(x *sqlparser.ComparisonExpr, injector *injector) (*compilerResult, error) {
	left, err := w.resolve(x.Left, injector)
	if err != nil {
		return nil, err
	}
	right, err := w.resolve(x.Right, injector)
	if err != nil {
		return nil, err
	}
	return &compilerResult{
		Content:              left.Content + " " + x.Operator + " " + right.Content,
		OriginalContent:      left.OriginalContent + " " + x.Operator + " " + right.OriginalContent,
		Args:                 append(left.Args, right.Args...),
		Fields:               left.Fields.merge(right.Fields),
		selectedExprs:        left.selectedExprs.merge(right.selectedExprs),
		selectedExprsReverse: right.selectedExprsReverse.merge(left.selectedExprsReverse),
	}, nil
}
