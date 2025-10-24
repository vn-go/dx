package sql

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

// expCmp.binary.go
func (e *expCmp) binary(leftExpr sqlparser.Expr, rightExpr sqlparser.Expr, operator string, injector *injector, cmpType CMP_TYP, selectedExprsReverse dictionaryFields) (*compilerResult, error) {
	left, err := e.resolve(leftExpr, injector, cmpType, selectedExprsReverse)
	if err != nil {
		return nil, err
	}
	right, err := e.resolve(rightExpr, injector, cmpType, selectedExprsReverse)

	if err != nil {

		return nil, traceCompilerError(err, fmt.Sprintf("%s %s %s", contents.toText(leftExpr), operator, contents.toText(rightExpr)))
	}
	ret := &compilerResult{
		OriginalContent:      fmt.Sprintf("%s %s %s", left.OriginalContent, operator, right.OriginalContent),
		Content:              fmt.Sprintf("%s %s %s", left.Content, operator, right.Content),
		Args:                 append(left.Args, right.Args...),
		Fields:               left.Fields.merge(right.Fields),
		selectedExprs:        dictionaryFields{},
		selectedExprsReverse: *left.selectedExprsReverse.merge(right.selectedExprsReverse),
		IsExpression:         true,
		IsInAggregateFunc:    left.IsExpression || right.IsExpression,
	}
	return ret, nil
}
