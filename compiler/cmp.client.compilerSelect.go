package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type cmpSelectorFieldType struct {
	expr  string
	alias string
}

func (cmp *cmpSelectorType) resolevSelector(dialect types.Dialect, outputFields *map[string]string, n sqlparser.SelectExprs, selector string) (string, error) {
	// if _, ok := n.(sqlparser.StarExpr); ok {
	// 	return "", NewCompilerError(fmt.Sprintf("'%s' is invalid expession"))
	// }
	strFields := []string{}
	for _, x := range n {
		f, err := cmp.resolve(dialect, outputFields, x, selector)
		if err != nil {
			return "", err
		}
		//valOfX := reflect.ValueOf(x)
		// fieldOfAs := valOfX.Elem().FieldByName("As")
		// if fieldOfAs.IsValid() {
		// 	val := fieldOfAs.Interface()
		// 	if sqlIndent, ok := val.(sqlparser.ColIdent); ok {
		// 		alias := sqlIndent.String()
		// 		fmt.Println(alias)
		// 	}
		// 	//f:=reflect.ValueOf(string)
		// }
		(*outputFields)[strings.ToLower(f.alias)] = f.expr
		strFields = append(strFields, fmt.Sprintf("%s %s", f.expr, dialect.Quote(f.alias)))

	}
	return strings.Join(strFields, ","), nil
}
func (cmp *cmpSelectorType) resolve(dialect types.Dialect, outputFields *map[string]string, n sqlparser.SQLNode, selector string) (*cmpSelectorFieldType, error) {
	if x, ok := n.(*sqlparser.AliasedExpr); ok {
		ret, err := cmp.resolve(dialect, outputFields, x.Expr, selector)
		if err != nil {
			return nil, err
		}
		if x.As.IsEmpty() {
			return ret, nil
		}
		return &cmpSelectorFieldType{
			expr:  ret.expr,
			alias: x.As.String(),
		}, nil

	}
	if x, ok := n.(*sqlparser.ColName); ok {
		if f, ok := (*outputFields)[x.Name.Lowered()]; ok {
			if cmp.cmpType == C_FUNC { //if is  in compling func return field no alias
				return &cmpSelectorFieldType{
					expr: f,
				}, nil
			}
			return &cmpSelectorFieldType{
				expr:  f,
				alias: x.Name.String(),
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
		return &cmpSelectorFieldType{
			expr: ret,
		}, nil

	}
	if x, ok := n.(*sqlparser.SQLVal); ok {
		v := string(x.Val)
		if strings.HasPrefix(v, ":v") {

			//n := *nextArgIndex + argIndex

			return &cmpSelectorFieldType{
				expr: "?",
			}, nil

		} else {
			if x.Type == sqlparser.StrVal {
				return &cmpSelectorFieldType{
					expr: dialect.ToText(v),
				}, nil
				//return dialect.ToText(v), nil
			}
			if internal.Helper.IsString(v) {
				return &cmpSelectorFieldType{
					expr: dialect.ToText(v),
				}, nil
				//return dialect.ToText(v), nil
			} else if internal.Helper.IsBool(v) {
				return &cmpSelectorFieldType{
					expr: dialect.ToBool(v),
				}, nil
				//return dialect.ToBool(v), nil
			} else if internal.Helper.IsFloatNumber(v) {
				return &cmpSelectorFieldType{
					expr: v,
				}, nil
				//return v, nil
			} else if internal.Helper.IsNumber(v) {
				return &cmpSelectorFieldType{
					expr: v,
				}, nil
				//return v, nil
			} else {
				return nil, NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid value", v, selector))
			}

		}
	}
	if isDebugMode {
		panic(fmt.Sprintf("Not implement %T, see 'resolve' in %s", n, `compiler\cmp.client.compilerSelect.go`))
	} else {
		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid syntax", selector))
	}

}
func (cmp *cmpSelectorType) resolveFuncExpr(dialect types.Dialect, outputFields *map[string]string, x *sqlparser.FuncExpr, selector string) (string, error) {
	oldCmpTYpe := cmp.cmpType
	defer func() {
		cmp.cmpType = oldCmpTYpe
	}()
	cmp.cmpType = C_FUNC
	strArgs := []string{}
	if x.Name.Lowered() == "contains" {
		if len(x.Exprs) != 2 {
			return "", newCompilerError(fmt.Sprintf("%s require 2 args. expression is '%s", x.Name.String(), selector), ERR)
		}

		for _, e := range x.Exprs {
			ex, err := cmp.resolve(dialect, outputFields, e, selector)
			if err != nil {
				return "", err
			}
			strArgs = append(strArgs, ex.expr)
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
		ex, err := cmp.resolve(dialect, outputFields, e, selector)
		if err != nil {
			return "", err
		}
		strArgs = append(strArgs, ex.expr)
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
		for i, v := range dialectDelegateFunction.Args {

			if _, ok := (x.Exprs[i].(sqlparser.SQLNode)).(*sqlparser.SQLVal); ok {
				newArgs = append(newArgs, v)
			} else {
				newArgs = append(newArgs, "COALESCE("+v+",'')")
			}

		}
		return dialectDelegateFunction.FuncName + "(" + strings.Join(newArgs, ", ") + ")", nil
	}
	return dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")", nil
}
