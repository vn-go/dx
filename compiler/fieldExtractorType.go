package compiler

import (
	"fmt"
	"reflect"

	"github.com/vn-go/dx/sqlparser"
)

type fieldExttractorType struct {
}

func (f *fieldExttractorType) GetFieldAlais(node sqlparser.SQLNode, visited map[string]bool) []string {
	ret := []string{}
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
						ret = append(ret, sqlIndent.String())
						found = true
					}
					fmt.Println(val)
				}

				//
			}
			if !found {
				r := f.GetFieldAlais(x, visited)
				if r != nil {
					ret = append(ret, r...)
				}
			}
		}
		return ret
	}
	if n, ok := node.(*sqlparser.Select); ok {
		return f.GetFieldAlais(n.SelectExprs, visited)
	}
	if n, ok := node.(*sqlparser.AliasedExpr); ok {
		if !n.As.IsEmpty() {
			ret = append(ret, n.As.String())
		} else {
			r := f.GetFieldAlais(n.Expr, visited)
			if r != nil {
				ret = append(ret, r...)
			}
			return ret
		}
		return ret
	}
	if n, ok := node.(*sqlparser.ColName); ok {
		ret = append(ret, n.Name.String())
		return ret
	}

	panic(fmt.Sprintf("Not impletement %T,`%s`", node, `compiler\fieldExtractorType.go`))

}

var FieldExttractor = &fieldExttractorType{}
