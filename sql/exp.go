package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type CMP_TYP int

const (
	CMP_SELECT CMP_TYP = iota
	CMP_WHERE
	CMP_TYP_FUNC
)

type expCmp struct {
}

func (e *expCmp) resolve(node sqlparser.SQLNode, injector *injector, cmpType CMP_TYP) (*compilerResult, error) {
	switch x := node.(type) {
	case *sqlparser.AndExpr:
		left, err := e.resolve(x.Left, injector, cmpType)
		if err != nil {
			return nil, err
		}
		right, err := e.resolve(x.Right, injector, cmpType)
		if err != nil {
			return nil, err
		}
		return &compilerResult{
			OriginalContent:      fmt.Sprintf("%s AND %s", left.OriginalContent, right.OriginalContent),
			Content:              fmt.Sprintf("%s AND %s", left.Content, right.Content),
			Args:                 append(left.Args, right.Args...),
			Fields:               left.Fields.merge(right.Fields),
			selectedExprs:        dictionaryFields{},
			selectedExprsReverse: dictionaryFields{},
			IsExpression:         true,
		}, nil
	case *sqlparser.OrExpr:
		left, err := e.resolve(x.Left, injector, cmpType)
		if err != nil {
			return nil, err
		}
		right, err := e.resolve(x.Right, injector, cmpType)
		if err != nil {
			return nil, err
		}
		return &compilerResult{
			OriginalContent:      fmt.Sprintf("%s OR %s", left.OriginalContent, right.OriginalContent),
			Content:              fmt.Sprintf("%s OR %s", left.Content, right.Content),
			Args:                 append(left.Args, right.Args...),
			Fields:               left.Fields.merge(right.Fields),
			selectedExprs:        dictionaryFields{},
			selectedExprsReverse: dictionaryFields{},
			IsExpression:         true,
		}, nil
	case *sqlparser.ComparisonExpr:
		left, err := e.resolve(x.Left, injector, cmpType)
		if err != nil {
			return nil, err
		}
		right, err := e.resolve(x.Right, injector, cmpType)
		if err != nil {
			return nil, err
		}
		return &compilerResult{
			OriginalContent:      fmt.Sprintf("%s %s %s", left.OriginalContent, x.Operator, right.OriginalContent),
			Content:              fmt.Sprintf("%s %s %s", left.Content, x.Operator, right.Content),
			Args:                 append(left.Args, right.Args...),
			Fields:               left.Fields.merge(right.Fields),
			selectedExprs:        dictionaryFields{},
			selectedExprsReverse: dictionaryFields{},
			IsExpression:         true,
		}, nil
	case *sqlparser.BinaryExpr:
		left, err := e.resolve(x.Left, injector, cmpType)
		if err != nil {
			return nil, err
		}
		right, err := e.resolve(x.Right, injector, cmpType)
		if err != nil {
			return nil, err
		}
		return &compilerResult{
			OriginalContent:      fmt.Sprintf("%s %s %s", left.OriginalContent, x.Operator, right.OriginalContent),
			Content:              fmt.Sprintf("%s %s %s", left.Content, x.Operator, right.Content),
			Args:                 append(left.Args, right.Args...),
			Fields:               left.Fields.merge(right.Fields),
			selectedExprs:        dictionaryFields{},
			selectedExprsReverse: dictionaryFields{},
			nonAggregateFields:   *left.nonAggregateFields.merge(right.nonAggregateFields),
			IsExpression:         true,
		}, nil
	case *sqlparser.ColName:
		return selector.colName(x, injector)
	case *sqlparser.SQLVal:
		return params.sqlVal(x, injector)
	case *sqlparser.FuncExpr:

		if x.Name.String() == GET_PARAMS_FUNC || x.Name.String() == internal.FnMarkSpecialTextArgs {
			return params.funcExpr(x, injector)
		} else {
			return e.funcExpr(x, injector, cmpType)
		}
	case *sqlparser.AliasedExpr:
		return e.aliasedExpr(x, injector, cmpType)

	default:
		panic(fmt.Sprintf("unhandled node type %T. see  expCmp.resolve, file %s", x, `sql\where.comparisonExpr.go`))
	}

}

func (s expCmp) aliasedExpr(expr *sqlparser.AliasedExpr, injector *injector, cmpType CMP_TYP) (*compilerResult, error) {
	switch t := expr.Expr.(type) {
	case *sqlparser.ColName:
		return selector.colName(t, injector)
	case *sqlparser.BinaryExpr:

		ret, err := exp.resolve(t, injector, cmpType)
		if err != nil {
			return nil, err
		}
		if cmpType == CMP_SELECT {
			if expr.As.IsEmpty() {
				return nil, newCompilerError("'%s' require alias", ret.OriginalContent)
			}
		}
		if cmpType == CMP_SELECT {
			ret.Content += " " + injector.dialect.Quote(expr.As.String())
		}
		ret.selectedExprs[strings.ToLower(ret.Content)] = &dictionaryField{
			Expr:              ret.Content,
			Typ:               -1,
			Alias:             expr.As.String(),
			IsInAggregateFunc: ret.IsInAggregateFunc,
		}

		ret.selectedExprsReverse[strings.ToLower(expr.As.String())] = ret.selectedExprs[strings.ToLower(ret.Content)]

		return ret, nil
	case *sqlparser.FuncExpr:

		ret, err := exp.funcExpr(t, injector, cmpType)
		if err != nil {
			return nil, err
		}
		if cmpType == CMP_SELECT {
			if expr.As.IsEmpty() {
				return nil, newCompilerError("'%s' require alias", ret.OriginalContent)
			}
		}
		if cmpType == CMP_SELECT {
			ret.Content += " " + injector.dialect.Quote(expr.As.String())
		}
		return ret, nil
	case *sqlparser.SQLVal:
		return params.sqlVal(t, injector)
	default:
		panic(fmt.Sprintf("unimplemented: %T. See selectors.aliasedExpr, %s", t, `sql\selectors.aliasedExpr.go.go`))

	}

}

func (e *expCmp) funcExpr(expr *sqlparser.FuncExpr, injector *injector, cmpType CMP_TYP) (*compilerResult, error) {
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
	delegator := types.DialectDelegateFunction{
		FuncName:         expr.Name.String(),
		Args:             []string{},
		ArgTypes:         []sqlparser.ValType{},
		IsAggregate:      false,
		HandledByDialect: false,
	}
	originItems := []string{}

	for _, arg := range expr.Exprs {
		argResult, err := e.resolve(arg, injector, cmpType)
		if err != nil {
			return nil, err
		}
		ret.Fields.merge(argResult.Fields)
		delegator.Args = append(delegator.Args, argResult.Content)
		originItems = append(originItems, argResult.OriginalContent)
		ret.nonAggregateFields.merge(argResult.nonAggregateFields)

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

var exp = &expCmp{}
