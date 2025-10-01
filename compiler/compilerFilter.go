package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type compilerFilterType struct {
}

func (cmp *compilerFilterType) Resolve(dialect types.Dialect, strFilter string, numOfParams *int, fields map[string]string, n sqlparser.SQLNode) (string, error) {
	if x, ok := n.(*sqlparser.ComparisonExpr); ok {
		if _, ok := x.Left.(*sqlparser.SQLVal); ok {
			return "", NewCompilerError(fmt.Sprintf("'%s' is vallid expression", strFilter))
		}
		left, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Left)
		if err != nil {
			return "", err
		}
		right, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Right)
		if err != nil {
			return "", err
		}
		if left == right {
			return "", NewCompilerError(fmt.Sprintf("'%s' is vallid expression", strFilter))
		}
		return left + " " + x.Operator + " " + right, nil
	}
	if x, ok := n.(*sqlparser.BinaryExpr); ok {
		left, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Left)
		if err != nil {
			return "", err
		}
		right, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Right)
		if err != nil {
			return "", err
		}
		return left + " " + x.Operator + " " + right, nil
	}
	if x, ok := n.(*sqlparser.ColName); ok {
		if x.Name.String() == "yes" || x.Name.String() == "no" || x.Name.String() == "true" || x.Name.String() == "false" {
			return dialect.ToBool(x.Name.String()), nil
		}
		if v, ok := fields[strings.ToLower(x.Name.String())]; ok {
			return v, nil
		} else {
			strFields := []string{}
			for k := range fields {
				strFields = append(strFields, k)
			}
			return "", newCompilerError(fmt.Sprintf("'%s' is not in , [%s],please review '%s'", x.Name.String(), strings.Join(strFields, ","), strFilter), ERR)
		}
	}
	if x, ok := n.(*sqlparser.SQLVal); ok {
		v := string(x.Val)
		if strings.HasPrefix(v, ":v") {
			n := *numOfParams + 1
			numOfParams = &n
			//n := *nextArgIndex + argIndex

			return "?", nil

		} else {
			if x.Type == sqlparser.StrVal {
				return dialect.ToText(v), nil
			}
			if internal.Helper.IsString(v) {
				return dialect.ToText(v), nil
			} else if internal.Helper.IsBool(v) {
				return dialect.ToBool(v), nil
			} else if internal.Helper.IsFloatNumber(v) {
				return v, nil
			} else if internal.Helper.IsNumber(v) {
				return v, nil
			} else {
				return "", NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid value", v, strFilter))
			}

		}
	}
	if x, ok := n.(*sqlparser.AndExpr); ok {
		left, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Left)
		if err != nil {
			return "", err
		}
		right, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Right)
		if err != nil {
			return "", err
		}
		return left + " AND " + right, nil
	}
	if x, ok := n.(*sqlparser.OrExpr); ok {
		left, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Left)
		if err != nil {
			return "", err
		}
		right, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Right)
		if err != nil {
			return "", err
		}
		return left + " OR" + right, nil
	}
	if x, ok := n.(*sqlparser.NotExpr); ok {
		left, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Expr)
		if err != nil {
			return "", err
		}

		return "NOT " + left, nil
	}
	if x, ok := n.(*sqlparser.FuncExpr); ok {
		return cmp.ResolveFunc(dialect, strFilter, numOfParams, fields, x)
	}
	if x, ok := n.(*sqlparser.AliasedExpr); ok {
		return cmp.Resolve(dialect, strFilter, numOfParams, fields, x.Expr)
	}
	return "", newCompilerError(fmt.Sprintf("'%s' is invalid expression ", strFilter), ERR)
	//panic(fmt.Sprintf("not implement %T, see 'Resolve' in file '%s'", n, `compiler\compilerFilter.go`))
}
func (cmp *compilerFilterType) ResolveFunc(dialect types.Dialect, strFilter string, numOfParams *int, fields map[string]string, x *sqlparser.FuncExpr) (string, error) {
	strArgs := []string{}
	if x.Name.Lowered() == "contains" {
		if len(x.Exprs) != 2 {
			return "", newCompilerError(fmt.Sprintf("%s require 2 args. expression is '%s", x.Name.String(), strFilter), ERR)
		}

		for _, e := range x.Exprs {
			ex, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, e)
			if err != nil {
				return "", err
			}
			strArgs = append(strArgs, ex)
		}
		dialectDelegateFunction := types.DialectDelegateFunction{
			FuncName:         "CONCAT",
			Args:             []string{"'%'", strArgs[1], "'%'"},
			HandledByDialect: false,
		}
		ret, err := dialect.SqlFunction(&dialectDelegateFunction)
		if err != nil {

			return "", err
		}
		if dialectDelegateFunction.HandledByDialect {

			return ret, nil
		}
		return strArgs[0] + " LIKE " + dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")", nil
	}
	for _, e := range x.Exprs {
		ex, err := cmp.Resolve(dialect, strFilter, numOfParams, fields, e)
		if err != nil {
			return "", err
		}
		strArgs = append(strArgs, ex)
	}

	dialectDelegateFunction := types.DialectDelegateFunction{
		FuncName:         x.Name.String(),
		Args:             strArgs,
		HandledByDialect: false,
	}
	ret, err := dialect.SqlFunction(&dialectDelegateFunction)
	if err != nil {

		return "", err
	}
	if dialectDelegateFunction.HandledByDialect {

		return ret, nil
	}
	if x.Name.Lowered() == "concat" {
		newArgs := []string{}
		for _, x := range dialectDelegateFunction.Args {
			newArgs = append(newArgs, "COALESCE("+x+",''")
		}
		return dialectDelegateFunction.FuncName + "(" + strings.Join(newArgs, ", ") + ")", nil
	}
	return dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")", nil
}

var CompilerFilter = &compilerFilterType{}
