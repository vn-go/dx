package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type FieldExprTypeEnum int

const (
	FieldExprType_Field                 FieldExprTypeEnum = iota // = 0
	FieldExprType_Expression                                     // = 1
	FieldExprType_AggregateFunctionCall                          // = 2
)

type FieldSelect struct {
	Expr          string
	Alias         string
	FieldExprType FieldExprTypeEnum
	Args          internal.SqlArgs
	FieldStat     map[string]FieldExprTypeEnum
	OriginalExpr  string
	FieldMap      map[string]string
}
type FieldSelects []FieldSelect
type ResolevSelectorResult struct {
	StrSelectors string
	Selectors    FieldSelects
	Args         internal.SqlArgs
	// all text constant in query has double apostrophe whill be exract here
	ApostropheArg []string
}

func (c *FieldSelects) HasAggregateFunction() bool {
	for _, x := range *c {
		if x.FieldExprType == FieldExprType_AggregateFunctionCall {
			return true
		}
	}
	return false
}
func (cmp *cmpSelectorType) resolevSelector(dialect types.Dialect, outputFields *map[string]types.OutputExpr, n sqlparser.SelectExprs, selector string, args *internal.SqlArgs, startOf2ApostropheArgs, startOfSqlIndex int) (*ResolevSelectorResult, error) {

	strFields := []string{}
	selectors := []FieldSelect{}

	for _, x := range n {
		f, err := cmp.resolve(dialect, outputFields, x, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}

		selectors = append(selectors, *f)
		(*outputFields)[strings.ToLower(f.Alias)] = types.OutputExpr{
			SqlNode:   x,
			FieldName: f.Expr,
			Expr: types.FiedlExpression{
				ExprContent:          f.Expr,
				FieldMapNotInAggFunc: f.FieldMap,
			},
			IsInAggregateFunc: f.FieldExprType == FieldExprType_AggregateFunctionCall,
		}
		strFields = append(strFields, fmt.Sprintf("%s %s", f.Expr, dialect.Quote(f.Alias)))

	}

	ret := &ResolevSelectorResult{
		StrSelectors: strings.Join(strFields, ","),
		Selectors:    selectors,
	}

	return ret, nil
}

func (cmp *cmpSelectorType) resolve(dialect types.Dialect,
	outputFields *map[string]types.OutputExpr, n sqlparser.SQLNode,
	selector string, args *internal.SqlArgs, startOf2ApostropheArgs, startOfSqlIndex int) (*FieldSelect, error) {
	if x, ok := n.(*sqlparser.AliasedExpr); ok {
		ret, err := cmp.resolve(dialect, outputFields, x.Expr, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}

		if !x.As.IsEmpty() && ret.Alias == "" {
			ret.Alias = x.As.String()
		}
		return ret, nil

	}
	if x, ok := n.(*sqlparser.ColName); ok {
		originalField := x.Name.String()
		if !x.Qualifier.IsEmpty() {
			originalField = x.Qualifier.Name.String() + "." + originalField
		}
		if outputFields == nil {
			return &FieldSelect{
				Expr:         x.Name.String(),
				Alias:        x.Name.String(),
				OriginalExpr: originalField,
				FieldMap:     map[string]string{x.Name.Lowered(): originalField},
			}, nil
		}
		if f, ok := (*outputFields)[x.Name.Lowered()]; ok {
			if cmp.cmpType == C_FUNC { //if is  in compling func return field no alias

				return &FieldSelect{
					Expr:         f.Expr.ExprContent,
					FieldStat:    map[string]FieldExprTypeEnum{x.Name.Lowered(): FieldExprType_Field},
					OriginalExpr: originalField,
					FieldMap:     map[string]string{x.Name.Lowered(): f.Expr.ExprContent},
				}, nil
			}

			return &FieldSelect{
				Expr:         f.Expr.ExprContent,
				Alias:        x.Name.String(),
				OriginalExpr: originalField,
				FieldMap:     map[string]string{x.Name.Lowered(): originalField},
			}, nil
			//return f + " " + dialect.Quote(x.Name.String()), nil
		} else {
			fieldList := []string{}
			for k := range *outputFields {
				fieldList = append(fieldList, k)
			}
			return nil, NewCompilerError(fmt.Sprintf("'%s' was not found in '%s',expression is '%s'", x.Name.String(), strings.Join(fieldList, ","), cmp.originalSelector))
		}
	}
	if x, ok := n.(*sqlparser.FuncExpr); ok {
		if x.Name.String() == internal.FnMarkSpecialTextArgs && len(x.Exprs) == 1 {
			n := x.Exprs[0].(*sqlparser.AliasedExpr).Expr.(*sqlparser.SQLVal)
			index, err := internal.Helper.ToIntFormBytes(n.Val)
			if err != nil {
				return nil, fmt.Errorf("%s is not int value", string(n.Val))
			}
			indexInSql := len(*args) + startOfSqlIndex + 1
			*args = append(*args, internal.SqlArg{
				ParamType:   internal.PARAM_TYPE_2APOSTROPHE,
				IndexInSql:  indexInSql,
				IndexInArgs: index + startOf2ApostropheArgs,
			})
			return &FieldSelect{
				Expr:          dialect.ToParam(indexInSql, sqlparser.StrVal),
				FieldExprType: FieldExprType_Expression,
			}, nil
		}
		defer func() {
			cmp.cmpType = C_SELECT
		}()
		cmp.cmpType = C_FUNC
		argsInFunc := internal.SqlArgs{}
		ret, err := cmp.resolveFuncExpr(dialect, outputFields, x, selector, &argsInFunc, startOf2ApostropheArgs, startOfSqlIndex)
		*args = append(*args, argsInFunc...)
		if err != nil {
			return nil, err
		}
		fieldExprType := FieldExprType_Expression
		if ret.isAggFuncCall {
			fieldExprType = FieldExprType_AggregateFunctionCall
		}
		return &FieldSelect{
			Expr:          ret.expr,
			FieldExprType: fieldExprType,
			Args:          argsInFunc,
			OriginalExpr:  ret.originalFuncCall,
			FieldMap:      ret.fieldMap,
		}, nil

	}
	if x, ok := n.(*sqlparser.SQLVal); ok {
		return cmp.resolveValue(dialect, x, selector, args, startOf2ApostropheArgs, startOfSqlIndex)

	}
	if _, ok := n.(*sqlparser.StarExpr); ok {
		return nil, NewCompilerError(fmt.Sprintf("%s' in '%s' is invalid expression, use CountAll instead", "count(*)", cmp.originalSelector))

	}
	if x, ok := n.(*sqlparser.BinaryExpr); ok {
		left, err := cmp.resolve(dialect, outputFields, x.Left, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		right, err := cmp.resolve(dialect, outputFields, x.Right, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		return &FieldSelect{
			Expr:          left.Expr + " " + x.Operator + " " + right.Expr,
			FieldExprType: FieldExprType_Expression,
			FieldStat:     internal.UnionMap(left.FieldStat, right.FieldStat),
			OriginalExpr:  left.OriginalExpr + " " + x.Operator + " " + right.OriginalExpr,
		}, nil
	}
	if x, ok := n.(*sqlparser.AndExpr); ok {
		left, err := cmp.resolve(dialect, outputFields, x.Left, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		right, err := cmp.resolve(dialect, outputFields, x.Right, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		if left.Expr == right.Expr {
			return nil, NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid expression", left.OriginalExpr+" or "+right.OriginalExpr, cmp.originalSelector))
		}
		return &FieldSelect{
			Expr:          left.Expr + " AND " + right.Expr,
			FieldExprType: FieldExprType_Expression,
			FieldStat:     internal.UnionMap(left.FieldStat, right.FieldStat),
			OriginalExpr:  left.OriginalExpr + " and " + right.OriginalExpr,
		}, nil
	}
	if x, ok := n.(*sqlparser.OrExpr); ok {
		left, err := cmp.resolve(dialect, outputFields, x.Left, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		right, err := cmp.resolve(dialect, outputFields, x.Right, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		if left.Expr == right.Expr {
			return nil, NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid expression", left.OriginalExpr+" or "+right.OriginalExpr, selector))
		}
		return &FieldSelect{
			Expr:          left.Expr + " OR " + right.Expr,
			FieldExprType: FieldExprType_Expression,
			FieldStat:     internal.UnionMap(left.FieldStat, right.FieldStat),
			OriginalExpr:  left.OriginalExpr + " or " + right.OriginalExpr,
		}, nil
	}
	if x, ok := n.(*sqlparser.ComparisonExpr); ok {
		left, err := cmp.resolve(dialect, outputFields, x.Left, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		right, err := cmp.resolve(dialect, outputFields, x.Right, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		return &FieldSelect{
			Expr:          left.Expr + " " + x.Operator + " " + right.Expr,
			FieldExprType: FieldExprType_Expression,
			FieldStat:     internal.UnionMap(left.FieldStat, right.FieldStat),
			OriginalExpr:  left.OriginalExpr + " " + x.Operator + " " + right.OriginalExpr,
		}, nil
	}
	if isDebugMode {
		panic(fmt.Sprintf("Not implement %T, see 'resolve' in %s", n, `compiler\cmp.client.compilerSelect.go`))
	} else {
		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid syntax", selector))
	}

}

type resolveFuncExprResult struct {
	expr             string
	isAggFuncCall    bool
	fieldStats       map[string]FieldExprTypeEnum
	originalFuncCall string
	fieldMap         map[string]string
}

func (cmp *cmpSelectorType) resolveFuncExpr(dialect types.Dialect,
	outputFields *map[string]types.OutputExpr, x *sqlparser.FuncExpr,
	selector string, args *internal.SqlArgs, startOf2ApostropheArgs, startOfSqlIndex int) (*resolveFuncExprResult, error) {
	oldCmpTYpe := cmp.cmpType
	defer func() {
		cmp.cmpType = oldCmpTYpe
	}()
	cmp.cmpType = C_FUNC
	strArgs := []string{}

	if x.Name.Lowered() == "contains" {
		if len(x.Exprs) != 2 {
			return nil, newCompilerError(fmt.Sprintf("%s require 2 args. expression is '%s", x.Name.String(), selector), ERR)
		}

		for _, e := range x.Exprs {
			ex, err := cmp.resolve(dialect, outputFields, e, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
			if err != nil {
				return nil, err
			}
			strArgs = append(strArgs, ex.Expr)
		}
		dialectDelegateFunction := types.DialectDelegateFunction{
			FuncName:         "CONCAT",
			Args:             []string{"'%'", strArgs[1], "'%'"},
			HandledByDialect: false,
		}
		ret, err := dialect.SqlFunction(&dialectDelegateFunction)
		if err != nil {

			return nil, err
		}
		if dialectDelegateFunction.HandledByDialect {
			return &resolveFuncExprResult{
				expr:          ret,
				isAggFuncCall: dialectDelegateFunction.IsAggregate,
			}, nil

			//return ret, nil
		}
		expr := strArgs[0] + " LIKE " + dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")"
		return &resolveFuncExprResult{
			expr: expr,
		}, nil
	}
	fieldStats := map[string]FieldExprTypeEnum{}
	originalArgs := []string{}
	fieldMap := map[string]string{}
	for _, e := range x.Exprs {
		ex, err := cmp.resolve(dialect, outputFields, e, selector, args, startOf2ApostropheArgs, startOfSqlIndex)
		if err != nil {
			return nil, err
		}
		originalArgs = append(originalArgs, ex.OriginalExpr)
		fieldStats = internal.UnionMap(fieldStats, ex.FieldStat)

		strArgs = append(strArgs, ex.Expr)
		fieldMap = internal.UnionMap(fieldMap, ex.FieldMap)
	}
	originalFuncCall := fmt.Sprintf("%s(%s)", x.Name, strings.Join(originalArgs, ","))
	dialectDelegateFunction := types.DialectDelegateFunction{
		FuncName:         x.Name.String(),
		Args:             strArgs,
		HandledByDialect: false,
	}
	ret, err := dialect.SqlFunction(&dialectDelegateFunction)
	if dialectDelegateFunction.IsAggregate {
		if cmp.aggregateExpr == nil {
			cmp.aggregateExpr = make(map[string]bool)
		}
		for _, x := range strArgs {
			if _, ok := cmp.aggregateExpr[x]; !ok {
				cmp.aggregateExpr[x] = true //mark all expr ar aggrgate for group by calculating
			}

		}

	}
	if err != nil {

		return nil, err
	}
	if dialectDelegateFunction.HandledByDialect {
		if dialectDelegateFunction.IsAggregate {
			fieldMap = map[string]string{}
		}
		return &resolveFuncExprResult{
			expr:             ret,
			isAggFuncCall:    dialectDelegateFunction.IsAggregate,
			fieldStats:       fieldStats,
			originalFuncCall: originalFuncCall,
			fieldMap:         fieldMap,
		}, nil

		//return ret, nil
	}
	if x.Name.Lowered() == "concat" {
		newArgs := []string{}
		for i, v := range dialectDelegateFunction.Args {

			if _, ok := (x.Exprs[i].(sqlparser.SQLNode)).(*sqlparser.SQLVal); ok {
				newArgs = append(newArgs, v)
			} else {
				newArgs = append(newArgs, "COALESCE("+v+",'')")
			}

		}
		//ret:=
		return &resolveFuncExprResult{
			expr:             dialectDelegateFunction.FuncName + "(" + strings.Join(newArgs, ", ") + ")",
			isAggFuncCall:    dialectDelegateFunction.IsAggregate,
			fieldStats:       fieldStats,
			originalFuncCall: originalFuncCall,
			fieldMap:         fieldMap,
		}, nil
	}
	if dialectDelegateFunction.IsAggregate {
		fieldMap = map[string]string{}
	}
	return &resolveFuncExprResult{
		expr:             dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")",
		isAggFuncCall:    dialectDelegateFunction.IsAggregate,
		fieldStats:       fieldStats,
		originalFuncCall: originalFuncCall,
		fieldMap:         fieldMap,
	}, nil
	//return dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")", nil
}
