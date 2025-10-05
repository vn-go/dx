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
type CompilerFilterTypeResult struct {
	expr      string
	fieldExpr string
	fields    map[string]string
}

func (c *CompilerFilterTypeResult) GetExpr() string {
	return c.expr
}
func (c *CompilerFilterTypeResult) GetFieldExpr() string {
	return c.fieldExpr
}
func (c *CompilerFilterTypeResult) GetFields() map[string]string {
	return c.fields
}
func (cmp *compilerFilterType) Resolve(dialect types.Dialect, strFilter string, fields map[string]string, n sqlparser.SQLNode) (*CompilerFilterTypeResult, error) {
	if x, ok := n.(*sqlparser.ComparisonExpr); ok {
		if _, ok := x.Left.(*sqlparser.SQLVal); ok {
			return nil, NewCompilerError(fmt.Sprintf("'%s' is vallid expression", strFilter))
		}
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
		if err != nil {
			return nil, err
		}
		if left == right {
			return nil, NewCompilerError(fmt.Sprintf("'%s' is vallid expression", strFilter))
		}
		expr := left.expr + " " + x.Operator + " " + right.expr
		fieldExpr := left.fieldExpr + " " + x.Operator + " " + right.fieldExpr
		fieldsSelected := map[string]string{}
		if left.fields != nil {
			for k, v := range left.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}

			}
		}
		if right.fields != nil {
			for k, v := range right.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}
			}
		}
		return &CompilerFilterTypeResult{
			expr:      expr,
			fieldExpr: fieldExpr,
			fields:    fieldsSelected,
		}, nil
	}
	if x, ok := n.(*sqlparser.BinaryExpr); ok {
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
		if err != nil {
			return nil, err
		}
		expr := left.expr + " " + x.Operator + " " + right.expr
		fieldsSelected := map[string]string{}
		if left.fields != nil {
			for k, v := range left.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}

			}
		}
		if right.fields != nil {
			for k, v := range right.fields {
				if _, ok := fields[k]; !ok {
					fieldsSelected[k] = v
				}
			}
		}

		return &CompilerFilterTypeResult{
			expr:   expr,
			fields: fieldsSelected,
		}, nil

	}
	if x, ok := n.(*sqlparser.ColName); ok {
		if x.Name.String() == "yes" || x.Name.String() == "no" || x.Name.String() == "true" || x.Name.String() == "false" {
			return &CompilerFilterTypeResult{
				expr: dialect.ToBool(x.Name.String()),
			}, nil

		}
		if v, ok := fields[strings.ToLower(x.Name.String())]; ok {
			return &CompilerFilterTypeResult{
				expr:      v,
				fieldExpr: dialect.Quote(x.Name.String()),
				fields:    map[string]string{strings.ToLower(x.Name.String()): x.Name.String()},
			}, nil

		} else {
			strFields := []string{}
			for k := range fields {
				strFields = append(strFields, k)
			}

			return nil, newCompilerError(fmt.Sprintf("'%s' is not in , [%s],please review '%s'", x.Name.String(), strings.Join(strFields, ","), strFilter), ERR)
		}
	}
	if x, ok := n.(*sqlparser.SQLVal); ok {
		v := string(x.Val)
		if strings.HasPrefix(v, ":v") {

			//n := *nextArgIndex + argIndex
			return &CompilerFilterTypeResult{
				expr:      "?",
				fieldExpr: "?",
			}, nil

		} else {
			if x.Type == sqlparser.StrVal {
				return &CompilerFilterTypeResult{
					expr:      dialect.ToText(v),
					fieldExpr: dialect.ToText(v),
				}, nil

			}
			if internal.Helper.IsString(v) {
				return &CompilerFilterTypeResult{
					expr:      dialect.ToText(v),
					fieldExpr: dialect.ToText(v),
				}, nil

			} else if internal.Helper.IsBool(v) {
				return &CompilerFilterTypeResult{
					expr:      dialect.ToBool(v),
					fieldExpr: dialect.ToBool(v),
				}, nil

			} else if internal.Helper.IsFloatNumber(v) {
				return &CompilerFilterTypeResult{
					expr:      v,
					fieldExpr: v,
				}, nil

			} else if internal.Helper.IsNumber(v) {
				return &CompilerFilterTypeResult{
					expr:      v,
					fieldExpr: v,
				}, nil

			} else {
				return nil, NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid value", v, strFilter))
			}

		}
	}
	if x, ok := n.(*sqlparser.AndExpr); ok {
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
		if err != nil {
			return nil, err
		}

		fieldsSelected := map[string]string{}
		if left.fields != nil {
			for k, v := range left.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}

			}
		}
		if right.fields != nil {
			for k, v := range right.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}
			}
		}

		return &CompilerFilterTypeResult{
			expr:      left.expr + " AND " + right.expr,
			fieldExpr: left.fieldExpr + " AND " + right.fieldExpr,
			fields:    fieldsSelected,
		}, nil

	}
	if x, ok := n.(*sqlparser.OrExpr); ok {
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
		if err != nil {
			return nil, err
		}
		fieldsSelected := map[string]string{}
		if left.fields != nil {
			for k, v := range left.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}

			}
		}
		if right.fields != nil {
			for k, v := range right.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}
			}
		}

		return &CompilerFilterTypeResult{
			expr:      left.expr + " OR " + right.expr,
			fieldExpr: left.fieldExpr + " OR " + right.fieldExpr,
			fields:    fieldsSelected,
		}, nil

	}
	if x, ok := n.(*sqlparser.NotExpr); ok {
		fieldsSelected := map[string]string{}
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Expr)
		if err != nil {
			return nil, err
		}
		if left.fields != nil {
			for k, v := range left.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldsSelected[k] = v
				}
			}
		}

		return &CompilerFilterTypeResult{
			expr:      "NOT " + left.expr,
			fieldExpr: "NOT " + left.fieldExpr,
			fields:    fieldsSelected,
		}, nil

	}
	if x, ok := n.(*sqlparser.FuncExpr); ok {
		return cmp.ResolveFunc(dialect, strFilter, fields, x)
	}
	if x, ok := n.(*sqlparser.AliasedExpr); ok {
		return cmp.Resolve(dialect, strFilter, fields, x.Expr)
	}
	if isDebugMode {
		panic(fmt.Sprintf("not implement %T, see 'Resolve' in file '%s'", n, `compiler\compilerFilter.go`))
	}
	return nil, newCompilerError(fmt.Sprintf("'%s' is invalid expression ", strFilter), ERR)

}
func (cmp *compilerFilterType) ResolveFunc(dialect types.Dialect, strFilter string, fields map[string]string, x *sqlparser.FuncExpr) (*CompilerFilterTypeResult, error) {
	strArgs := []string{}
	if x.Name.Lowered() == "contains" {
		if len(x.Exprs) != 2 {
			return nil, newCompilerError(fmt.Sprintf("%s require 2 args. expression is '%s", x.Name.String(), strFilter), ERR)
		}
		fieldsSelected := map[string]string{}
		for _, e := range x.Exprs {
			ex, err := cmp.Resolve(dialect, strFilter, fields, e)
			if err != nil {
				return nil, err
			}
			if ex.fields != nil {
				for k, v := range ex.fields {
					if _, ok := fieldsSelected[k]; !ok {
						fieldsSelected[k] = v
					}

				}
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

			return nil, err
		}
		if dialectDelegateFunction.HandledByDialect {

			return &CompilerFilterTypeResult{
				expr:   ret,
				fields: fieldsSelected,
			}, nil
		}
		ret = strArgs[0] + " LIKE " + dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")"
		return &CompilerFilterTypeResult{
			expr:   ret,
			fields: fieldsSelected,
		}, nil
	}
	fieldsSelected := map[string]string{}
	for _, e := range x.Exprs {
		fieldsSelected = map[string]string{}
		ex, err := cmp.Resolve(dialect, strFilter, fields, e)
		if err != nil {
			return nil, err
		}
		if ex.fields != nil {
			for k, v := range ex.fields {
				if _, ok := fieldsSelected[k]; !ok {
					fields[k] = v
				}
			}
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

		return nil, err
	}
	if dialectDelegateFunction.HandledByDialect {

		return &CompilerFilterTypeResult{
			expr:   ret,
			fields: fieldsSelected,
		}, nil
	}
	if x.Name.Lowered() == "concat" {
		newArgs := []string{}
		for _, x := range dialectDelegateFunction.Args {
			newArgs = append(newArgs, "COALESCE("+x+",'')")
		}
		ret := dialectDelegateFunction.FuncName + "(" + strings.Join(newArgs, ", ") + ")"
		return &CompilerFilterTypeResult{
			expr:   ret,
			fields: fieldsSelected,
		}, nil
	}
	ret = dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")"
	return &CompilerFilterTypeResult{
		expr:   ret,
		fields: fieldsSelected,
	}, nil
}

var CompilerFilter = &compilerFilterType{}
