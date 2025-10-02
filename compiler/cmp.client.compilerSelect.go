package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type CmpSelectorFieldType struct {
	Expr          string
	Alias         string
	IsAggFuncCall bool
}
type ResolevSelectorResult struct {
	StrSelectors string
	Selectors    []CmpSelectorFieldType
}

func (cmp *cmpSelectorType) resolevSelector(dialect types.Dialect, outputFields *map[string]string, n sqlparser.SelectExprs, selector string) (*ResolevSelectorResult, error) {
	// if _, ok := n.(sqlparser.StarExpr); ok {
	// 	return "", NewCompilerError(fmt.Sprintf("'%s' is invalid expession"))
	// }
	strFields := []string{}
	selectors := []CmpSelectorFieldType{}
	// exprMap := map[string]string{}
	// exprMapRevert := map[string]string{}

	// fieldsInAggFunc := map[string]string{}
	//fieldsInGroup := map[string]string{}
	// exprAgg := map[string]string{}
	// exprAggRevert := map[string]string{}
	for _, x := range n {
		f, err := cmp.resolve(dialect, outputFields, x, selector)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, *f)
		(*outputFields)[strings.ToLower(f.Alias)] = f.Expr
		strFields = append(strFields, fmt.Sprintf("%s %s", f.Expr, dialect.Quote(f.Alias)))
		// exprMap[f.alias[1:len(f.alias)-1]] = f.expr
		// exprMapRevert[f.expr] = f.expr
		// if !f.isAggFuncCall {
		// 	if _, ok := cmp.aggregateExpr[f.expr]; !ok {
		// 		fieldsInGroup[f.alias] = f.expr
		// 	}
		// } else {
		// 	exprAgg[strings.ToLower(f.alias)] = f.expr
		// 	exprAggRevert[strings.ToLower(f.expr)] = f.alias

		// }

	}
	// if cmp.aggregateExpr != nil {
	// 	for k := range cmp.aggregateExpr {
	// 		fieldsInAggFunc[strings.ToLower(k)] = k

	// 	}

	// }
	ret := &ResolevSelectorResult{
		StrSelectors: strings.Join(strFields, ","),
		Selectors:    selectors,
		// ExprMap:       exprMap,
		// ExprMapRevert: exprMapRevert,
		// FieldsInGroup: fieldsInGroup,
		// ExprAgg:       exprAgg,
		// ExprAggRevert: exprAggRevert,
	}

	return ret, nil
}
func (cmp *cmpSelectorType) resolve(dialect types.Dialect, outputFields *map[string]string, n sqlparser.SQLNode, selector string) (*CmpSelectorFieldType, error) {
	if x, ok := n.(*sqlparser.AliasedExpr); ok {
		ret, err := cmp.resolve(dialect, outputFields, x.Expr, selector)
		if err != nil {
			return nil, err
		}

		if !x.As.IsEmpty() && ret.Alias == "" {
			ret.Alias = x.As.String()
		}
		return ret, nil

		// return &CmpSelectorFieldType{
		// 	Expr:  ret.expr,
		// 	Alias: x.As.String(),
		// 	IsAggFuncCall: ,
		// }, nil

	}
	if x, ok := n.(*sqlparser.ColName); ok {
		if f, ok := (*outputFields)[x.Name.Lowered()]; ok {
			if cmp.cmpType == C_FUNC { //if is  in compling func return field no alias
				return &CmpSelectorFieldType{
					Expr: f,
				}, nil
			}

			return &CmpSelectorFieldType{
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
		ret, err := cmp.resolveFuncExpr(dialect, outputFields, x, selector)
		if err != nil {
			return nil, err
		}
		return &CmpSelectorFieldType{
			Expr:          ret.expr,
			IsAggFuncCall: ret.isAggFuncCall,
		}, nil

	}
	if x, ok := n.(*sqlparser.SQLVal); ok {
		v := string(x.Val)
		if strings.HasPrefix(v, ":v") {

			//n := *nextArgIndex + argIndex

			return &CmpSelectorFieldType{
				Expr: "?",
			}, nil

		} else {
			if x.Type == sqlparser.StrVal {
				return &CmpSelectorFieldType{
					Expr: dialect.ToText(v),
				}, nil
				//return dialect.ToText(v), nil
			}
			if internal.Helper.IsString(v) {
				return &CmpSelectorFieldType{
					Expr: dialect.ToText(v),
				}, nil
				//return dialect.ToText(v), nil
			} else if internal.Helper.IsBool(v) {
				return &CmpSelectorFieldType{
					Expr: dialect.ToBool(v),
				}, nil
				//return dialect.ToBool(v), nil
			} else if internal.Helper.IsFloatNumber(v) {
				return &CmpSelectorFieldType{
					Expr: v,
				}, nil
				//return v, nil
			} else if internal.Helper.IsNumber(v) {
				return &CmpSelectorFieldType{
					Expr: v,
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

func (cmp *cmpSelectorType) resolveFuncExpr(dialect types.Dialect, outputFields *map[string]string, x *sqlparser.FuncExpr, selector string) (*resolveFuncExprResult, error) {
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
			ex, err := cmp.resolve(dialect, outputFields, e, selector)
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
		ex, err := cmp.resolve(dialect, outputFields, e, selector)
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
