package sql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// exp.detectIsConcatFunc.go
func (e *expCmp) detectIsConcatFunc(expr sqlparser.Expr, fields dictionaryFields) sqlparser.Expr {
	if fx, ok := expr.(*sqlparser.BinaryExpr); ok {
		if e.areAllOperatorPlus(fx, fields) {

			nodes, ok := e.extractAllArgNodes(fx, fields)
			if !ok {
				return expr
			}

			return e.convertConcatFunc(fx, nodes)
		}
	}
	return expr
}

func (e *expCmp) convertConcatFunc(fx *sqlparser.BinaryExpr, nodes []sqlparser.SQLNode) sqlparser.Expr {
	args := []sqlparser.SelectExpr{}
	for _, node := range nodes {
		args = append(args, &sqlparser.AliasedExpr{Expr: node.(sqlparser.Expr)})
	}
	ret := &sqlparser.FuncExpr{
		Name:  sqlparser.NewColIdent("concat"),
		Exprs: args,
	}
	fmt.Println(smartier.ToText(ret))
	return ret
}

func (e *expCmp) extractAllArgNodes(node sqlparser.SQLNode, fields dictionaryFields) ([]sqlparser.SQLNode, bool) {
	ret := []sqlparser.SQLNode{}
	switch expr := node.(type) {
	case *sqlparser.BinaryExpr:
		left, ok := e.extractAllArgNodes(expr.Left, fields)
		if !ok {
			return nil, false
		}
		right, ok := e.extractAllArgNodes(expr.Right, fields)
		if !ok {
			return nil, false
		}
		return append(left, right...), true

	case *sqlparser.ColName:
		key := strings.ToLower(fmt.Sprintf("%s.%s", expr.Qualifier.Name.String(), expr.Name.String()))
		if f, ok := fields[key]; ok {
			if f.EntityField.Field.Type == reflect.TypeFor[string]() || f.EntityField.Field.Type == reflect.TypeFor[*string]() {
				ret = append(ret, node)
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}

	case *sqlparser.SQLVal:
		ret = append(ret, node)
	case *sqlparser.FuncExpr:
		if _, ok := textFunc[expr.Name.Lowered()]; ok {
			ret = append(ret, node)
		}
		return nil, false

	default:
		panic(fmt.Sprintf("unexpected type %T, ref expCmp.extractAllArgNodes, file %s", expr, `sql\exp.detectIsConcatFunc.go`))
	}
	return ret, true
}

func (e *expCmp) allColsOrParamsIsText(fx *sqlparser.BinaryExpr, fields dictionaryFields) bool {
	return e.allColsOrParamsIsTextExpr(fx.Left, fields) && e.allColsOrParamsIsTextExpr(fx.Right, fields)
}

var textFunc = map[string]bool{
	"concat": true,
	"text":   true,
	"upper":  true,
	"lower":  true,
	"left":   true,
	"right":  true,
	"substr": true,
}

func (e *expCmp) allColsOrParamsIsTextExpr(expr sqlparser.Expr, fields dictionaryFields) bool {
	switch expr := expr.(type) {
	case *sqlparser.ColName:
		if f, ok := fields[expr.Name.Lowered()]; ok {
			return f.EntityField.Field.Type == reflect.TypeFor[string]() || f.EntityField.Field.Type == reflect.TypeFor[*string]()
		}
		return false
	case *sqlparser.SQLVal:
		return true
	case *sqlparser.FuncExpr:
		if _, ok := textFunc[expr.Name.Lowered()]; ok {
			return true
		}
		return false
	case *sqlparser.BinaryExpr:
		return e.allColsOrParamsIsText(expr, fields)
	default:
		panic(fmt.Sprintf("unexpected type %T, ref expCmp.allColsOrParamsIsTextExpr, file %s", expr, `sql\exp.detectIsConcatFunc.go`))
	}
}
func (e *expCmp) areAllOperatorPlus(fx *sqlparser.BinaryExpr, fields dictionaryFields) bool {
	if fx.Operator == "+" {
		return true
	}
	left, okLeft := fx.Left.(*sqlparser.BinaryExpr)
	if !okLeft {
		return false
	}
	right, okRight := fx.Right.(*sqlparser.BinaryExpr)
	if !okRight {
		return false
	}
	return e.areAllOperatorPlus(left, fields) && e.areAllOperatorPlus(right, fields)
}
