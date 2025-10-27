package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

// exp.funcExpr.go
func (e *expCmp) funcExpr(expr *sqlparser.FuncExpr, injector *injector, cmpType CMP_TYP, selectedExprsReverse dictionaryFields) (*compilerResult, error) {
	oldCmpType := cmpType
	defer func() {
		cmpType = oldCmpType
	}()
	cmpType = CMP_TYP_FUNC

	ret := &compilerResult{
		selectedExprs:        dictionaryFields{},
		selectedExprsReverse: dictionaryFields{},
		Fields:               refFields{},
		nonAggregateFields:   dictionaryFields{},
	}
	fName := expr.Name.String()
	if !expr.Qualifier.IsEmpty() {
		fName = expr.Qualifier.String() + "." + fName
	}
	delegator := types.DialectDelegateFunction{
		FuncName:         fName,
		Args:             []string{},
		ArgTypes:         []sqlparser.ValType{},
		IsAggregate:      false,
		HandledByDialect: false,
	}
	originItems := []string{}

	for _, arg := range expr.Exprs {
		argResult, err := e.resolve(arg, injector, cmpType, selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		ret.Fields.merge(argResult.Fields)
		delegator.Args = append(delegator.Args, argResult.Content)
		originItems = append(originItems, argResult.OriginalContent)
		ret.nonAggregateFields.merge(argResult.nonAggregateFields)
		ret.Fields.merge(argResult.Fields) // important: we need to get all field for data acess permission check
		ret.selectedExprsReverse.merge(argResult.selectedExprsReverse)
	}
	content, err := injector.dialect.SqlFunction(&delegator)
	if err != nil {
		return nil, err
	}

	ret.OriginalContent = fmt.Sprintf("%s(%s)", expr.Name.String(), strings.Join(originItems, ", "))
	if delegator.HandledByDialect {
		ret.Content = content
	} else {
		ret.Content = fmt.Sprintf("%s(%s)", expr.Name.String(), strings.Join(delegator.Args, ", "))
	}
	ret.IsInAggregateFunc = delegator.IsAggregate
	if delegator.IsAggregate {
		ret.nonAggregateFields = dictionaryFields{} // no need to keep non-aggregate fields
	}
	ret.IsExpression = true

	return ret, nil
	//panic(fmt.Sprintf("unhandled node type %s. see  expCmp.funcExpr, file %s", expr.Name.String(), `sql\where.comparisonExpr.go`))
}
