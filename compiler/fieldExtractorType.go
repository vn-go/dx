package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

type fieldExttractorType struct {
}

func (f *fieldExttractorType) GetFieldAlais(node sqlparser.SQLNode, visited map[string]bool, isSubQuery bool) (map[string]types.OutputExpr, error) {

	ret := map[string]types.OutputExpr{}
	if n, ok := node.(sqlparser.SelectExprs); ok {
		for _, x := range n {
			if _, ok := x.(*sqlparser.StarExpr); ok {
				continue
			}
			typ := reflect.TypeOf(x)
			val := reflect.ValueOf(x)
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
				val = val.Elem()
			}
			found := false
			if fx, ok := typ.FieldByName("As"); ok {
				//<sqlparser.ColIdent Value>
				val := val.FieldByIndex(fx.Index).Interface()
				if sqlIndent, ok := val.(sqlparser.ColIdent); ok {
					if sqlIndent.String() != "" {
						if _, ok := visited[strings.ToLower(sqlIndent.String())]; ok {
							return nil, newCompilerError(fmt.Sprintf("Duplicate select field,'%s'", sqlIndent.String()), ERR)
						}
						ret[strings.ToLower(sqlIndent.String())] = types.OutputExpr{
							SqlNode:   x,
							FieldName: sqlIndent.String(),
							Expr:      sqlIndent.String(),
						}

						visited[strings.ToLower(sqlIndent.String())] = true
						found = true
					}

				}

				//
			}
			if !found {
				r, err := f.GetFieldAlais(x, visited, isSubQuery)
				if err != nil {
					return nil, err
				}
				if r != nil {
					for k, v := range r {
						if _, ok := ret[k]; ok {
							return nil, newCompilerError(fmt.Sprintf("Duplicate select field,'%s'", v.FieldName), ERR)
						}
						ret[k] = v
					}

				}
			}
		}
		return ret, nil
	}
	if n, ok := node.(*sqlparser.Select); ok {
		return f.GetFieldAlais(n.SelectExprs, visited, isSubQuery)
	}
	if n, ok := node.(*sqlparser.AliasedExpr); ok {
		if !n.As.IsEmpty() {
			if _, ok := visited[strings.ToLower(n.As.String())]; ok {
				return nil, newCompilerError(fmt.Sprintf("Duplicate select field,'%s'", n.As.String()), ERR)
			}
			ret[strings.ToLower(n.As.String())] = types.OutputExpr{
				SqlNode:   node,
				FieldName: n.As.String(),
			}

			visited[strings.ToLower(n.As.String())] = true
		} else {
			r, err := f.GetFieldAlais(n.Expr, visited, isSubQuery)
			if err != nil {
				return nil, err
			}
			if r != nil {

				for k, v := range r {
					if _, ok := ret[k]; ok {
						return nil, newCompilerError(fmt.Sprintf("Duplicate select field,'%s'", v.FieldName), ERR)
					}
					ret[k] = v
				}
			}
			return ret, nil
		}
		return ret, nil
	}
	if n, ok := node.(*sqlparser.ColName); ok {
		if _, ok := visited[strings.ToLower(n.Name.String())]; ok {
			return nil, newCompilerError(fmt.Sprintf("Duplicate select field,'%s'", n.Name.String()), ERR)
		}
		ret[strings.ToLower(n.Name.String())] = types.OutputExpr{
			SqlNode:   node,
			FieldName: n.Name.String(),
			Expr:      n.Name.String(),
		}

		visited[strings.ToLower(n.Name.String())] = true
		return ret, nil
	}
	if n, ok := node.(*sqlparser.Union); ok {

		left, err := f.GetFieldAlais(n.Left, map[string]bool{}, isSubQuery)
		if err != nil {
			return nil, err
		}
		right, err := f.GetFieldAlais(n.Right, map[string]bool{}, isSubQuery)
		if err != nil {
			return nil, err
		}
		for k, v := range right {
			left[k] = v
		}
		return left, nil
	}
	panic(fmt.Sprintf("Not impletement %T,`%s`", node, `compiler\fieldExtractorType.go`))

}

var FieldExttractor = &fieldExttractorType{}
