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
	CMP_ORDER_BY
	CMP_JOIN
	CMP_SUBQUERY
	CMP_GROUP
)

type expCmp struct {
}

func (e *expCmp) resolve(node sqlparser.SQLNode, injector *injector, cmpType CMP_TYP, selectedExprsReverse dictionaryFields) (*compilerResult, error) {
	switch x := node.(type) {
	case *sqlparser.AndExpr:
		return e.binary(x.Left, x.Right, "AND", injector, cmpType, selectedExprsReverse)

	case *sqlparser.OrExpr:
		return e.binary(x.Left, x.Right, "OR", injector, cmpType, selectedExprsReverse)
	case *sqlparser.ComparisonExpr:
		ret, err := e.binary(x.Left, x.Right, x.Operator, injector, cmpType, selectedExprsReverse)
		return ret, err

	case *sqlparser.BinaryExpr:
		return e.binary(x.Left, x.Right, x.Operator, injector, cmpType, selectedExprsReverse)
	case *sqlparser.ColName:
		return selector.colName(x, injector, cmpType, selectedExprsReverse)
	case *sqlparser.SQLVal:
		return params.sqlVal(x, injector)
	case *sqlparser.FuncExpr:

		if x.Name.String() == GET_PARAMS_FUNC || x.Name.String() == internal.FnMarkSpecialTextArgs {
			return params.funcExpr(x, injector)
		} else {
			return e.funcExpr(x, injector, cmpType, selectedExprsReverse)
		}
	case *sqlparser.AliasedExpr:
		return e.aliasedExpr(x, injector, cmpType, selectedExprsReverse)
	case *sqlparser.NotExpr:
		fx, err := e.resolve(x.Expr, injector, cmpType, selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		fx.Content = "NOT " + fx.Content
		fx.OriginalContent = "NOT " + fx.OriginalContent
		return fx, nil

	default:
		panic(fmt.Sprintf("unhandled node type %T. see  expCmp.resolve, file %s", x, `sql\where.comparisonExpr.go`))
	}

}

func (s expCmp) aliasedExpr(expr *sqlparser.AliasedExpr, injector *injector, cmpType CMP_TYP, selectedExprsReverse dictionaryFields) (*compilerResult, error) {
	switch t := expr.Expr.(type) {
	case *sqlparser.ColName:
		return selector.colName(t, injector, cmpType, selectedExprsReverse)
	case *sqlparser.BinaryExpr:

		ret, err := exp.resolve(t, injector, cmpType, selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		if cmpType == CMP_SELECT {
			if expr.As.IsEmpty() {
				return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "'%s' require alias", ret.OriginalContent)
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
		if t.Name.String() == GET_PARAMS_FUNC || t.Name.String() == internal.FnMarkSpecialTextArgs {
			return params.funcExpr(t, injector)
		}

		ret, err := exp.funcExpr(t, injector, cmpType, selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		if cmpType == CMP_SELECT {
			if expr.As.IsEmpty() {
				return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "'%s' require alias", ret.OriginalContent)
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
	delegator := types.DialectDelegateFunction{
		FuncName:         expr.Name.String(),
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

var exp = &expCmp{}
