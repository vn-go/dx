package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type FieldSelect struct {
	Expr          string
	Alias         string
	IsAggFuncCall bool
	Args          []any
}
type ResolevSelectorResult struct {
	StrSelectors string
	Selectors    []FieldSelect
	Args         internal.SelectorTypesArgs
}

func (cmp *cmpSelectorType) resolevSelector(dialect types.Dialect, outputFields *map[string]string, n sqlparser.SelectExprs, selector string, args *[]any) (*ResolevSelectorResult, error) {

	strFields := []string{}
	selectors := []FieldSelect{}

	for _, x := range n {
		f, err := cmp.resolve(dialect, outputFields, x, selector, args)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, *f)
		(*outputFields)[strings.ToLower(f.Alias)] = f.Expr
		strFields = append(strFields, fmt.Sprintf("%s %s", f.Expr, dialect.Quote(f.Alias)))

	}

	ret := &ResolevSelectorResult{
		StrSelectors: strings.Join(strFields, ","),
		Selectors:    selectors,
	}

	return ret, nil
}
func (cmp *cmpSelectorType) resolve(dialect types.Dialect, outputFields *map[string]string, n sqlparser.SQLNode, selector string, args *[]any) (*FieldSelect, error) {
	if x, ok := n.(*sqlparser.AliasedExpr); ok {
		ret, err := cmp.resolve(dialect, outputFields, x.Expr, selector, args)
		if err != nil {
			return nil, err
		}

		if !x.As.IsEmpty() && ret.Alias == "" {
			ret.Alias = x.As.String()
		}
		return ret, nil

	}
	if x, ok := n.(*sqlparser.ColName); ok {
		if f, ok := (*outputFields)[x.Name.Lowered()]; ok {
			if cmp.cmpType == C_FUNC { //if is  in compling func return field no alias
				return &FieldSelect{
					Expr: f,
				}, nil
			}

			return &FieldSelect{
				Expr:  f,
				Alias: x.Name.String(),
			}, nil
			//return f + " " + dialect.Quote(x.Name.String()), nil
		} else {
			fieldList := []string{}
			for k := range *outputFields {
				fieldList = append(fieldList, k)
			}
			return nil, NewCompilerError(fmt.Sprintf("'%s' was not found in '%s',expression is '%s'", x.Name.String(), strings.Join(fieldList, ","), selector))
		}
	}
	if x, ok := n.(*sqlparser.FuncExpr); ok {
		defer func() {
			cmp.cmpType = C_SELECT
		}()
		cmp.cmpType = C_FUNC
		argsInFunc := []any{}
		ret, err := cmp.resolveFuncExpr(dialect, outputFields, x, selector, &argsInFunc)
		*args = append(*args, argsInFunc...)
		if err != nil {
			return nil, err
		}
		return &FieldSelect{
			Expr:          ret.expr,
			IsAggFuncCall: ret.isAggFuncCall,
			Args:          argsInFunc,
		}, nil

	}
	if x, ok := n.(*sqlparser.SQLVal); ok {
		v := string(x.Val)
		if strings.HasPrefix(v, ":v") {
			index, err := strconv.ParseInt(v[2:], 32, 32)
			if err != nil {
				return nil, NewCompilerError(fmt.Sprintf("%s is invalid expression", selector))
			}
			//n := *nextArgIndex + argIndex
			*args = append(*args, internal.QuestionArg{
				Index: int(index),
			})
			return &FieldSelect{
				Expr: "?",
			}, nil

		} else {
			if x.Type == sqlparser.StrVal {
				*args = append(*args, dialect.ToText(v))
				return &FieldSelect{
					Expr: "?",
				}, nil
				//return dialect.ToText(v), nil
			}
			if internal.Helper.IsString(v) {
				*args = append(*args, dialect.ToText(v))
				return &FieldSelect{
					Expr: "?",
				}, nil
				//return dialect.ToText(v), nil
			} else if internal.Helper.IsBool(v) {
				*args = append(*args, dialect.ToBool(v))
				return &FieldSelect{
					Expr: "?",
				}, nil
				//return dialect.ToBool(v), nil
			} else if internal.Helper.IsFloatNumber(v) {
				fx, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, NewCompilerError(fmt.Sprintf("%s is invalid expression", selector))
				}
				*args = append(*args, fx)
				return &FieldSelect{
					Expr: "?",
				}, nil
				//return v, nil
			} else if internal.Helper.IsNumber(v) {
				fx, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return nil, NewCompilerError(fmt.Sprintf("%s is invalid expression", selector))
				}
				*args = append(*args, fx)
				return &FieldSelect{
					Expr: "?",
				}, nil
				//return v, nil
			} else {
				return nil, NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid value", v, selector))
			}

		}
	}
	if _, ok := n.(*sqlparser.StarExpr); ok {
		return nil, NewCompilerError(fmt.Sprintf("%s' in '%s' is invalid expression, use CountAll instead", "count(*)", selector))

	}
	if isDebugMode {
		panic(fmt.Sprintf("Not implement %T, see 'resolve' in %s", n, `compiler\cmp.client.compilerSelect.go`))
	} else {
		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid syntax", selector))
	}

}

type resolveFuncExprResult struct {
	expr          string
	isAggFuncCall bool
}

func (cmp *cmpSelectorType) resolveFuncExpr(dialect types.Dialect, outputFields *map[string]string, x *sqlparser.FuncExpr, selector string, args *[]any) (*resolveFuncExprResult, error) {
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
			ex, err := cmp.resolve(dialect, outputFields, e, selector, args)
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
	for _, e := range x.Exprs {
		ex, err := cmp.resolve(dialect, outputFields, e, selector, args)
		if err != nil {
			return nil, err
		}
		strArgs = append(strArgs, ex.Expr)
	}

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
		return &resolveFuncExprResult{
			expr:          ret,
			isAggFuncCall: dialectDelegateFunction.IsAggregate,
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
			expr:          dialectDelegateFunction.FuncName + "(" + strings.Join(newArgs, ", ") + ")",
			isAggFuncCall: dialectDelegateFunction.IsAggregate,
		}, nil
	}
	return &resolveFuncExprResult{
		expr:          dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")",
		isAggFuncCall: dialectDelegateFunction.IsAggregate,
	}, nil
	//return dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")", nil
}
