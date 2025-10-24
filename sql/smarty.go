package sql

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type smarty struct {
}

func (s *smarty) from(selectStm *sqlparser.Select) string {

	dsSourceFunc := s.findFncExpr(selectStm, "from")
	if dsSourceFunc != nil {
		comparisonExprs := []sqlparser.SQLNode{}
		aliasedExpr := []sqlparser.SQLNode{}
		for _, expr := range dsSourceFunc.Exprs {
			if isNode[*sqlparser.ComparisonExpr](expr) {
				comparisonExprs = append(comparisonExprs, expr)
			} else if aliased, ok := expr.(*sqlparser.AliasedExpr); ok {
				aliasedExpr = append(aliasedExpr, aliased)
			}
		}

		if len(comparisonExprs) > 0 || len(aliasedExpr) > 0 {
			return s.convertToJoinTableExpr(comparisonExprs, aliasedExpr)
		}
	} else {

		return s.convertToTableExprs(selectStm)
	}
	return ""
}

func (s *smarty) findFncExpr(selectStm sqlparser.SQLNode, fncName string) *sqlparser.FuncExpr {
	fnCallName := fncName
	qualifierName := ""
	if strings.Contains(fncName, ".") {
		fnCallName = strings.ToLower(strings.Split(fncName, ".")[1])
		qualifierName = strings.ToLower(strings.Split(fncName, ".")[0])
	}
	switch t := selectStm.(type) {
	case *sqlparser.Select:
		for _, expr := range t.SelectExprs {
			if fn, ok := expr.(sqlparser.SQLNode).(*sqlparser.FuncExpr); ok {
				if fn.Name.Lowered() == fnCallName && strings.ToLower(fn.Qualifier.String()) == qualifierName {
					return fn
				}
			} else {
				if r := s.findFncExpr(expr.(sqlparser.SQLNode), fncName); r != nil {
					return r
				}
			}
		}
		return nil
	case *sqlparser.AliasedExpr:
		if fn, ok := t.Expr.(sqlparser.SQLNode).(*sqlparser.FuncExpr); ok {
			if fn.Name.Lowered() == fnCallName && strings.ToLower(fn.Qualifier.String()) == qualifierName {
				return fn
			}
		}
		return nil
	default:
		panic(fmt.Sprintf("unexpected type %T. ref smarty.findFncExpr", t))
	}
}
func isNode[T any](node sqlparser.SQLNode) bool {
	switch t := node.(type) {
	case T:
		return true
	case *sqlparser.AliasedExpr:
		return isNode[T](t.Expr)
	case *sqlparser.ColName:
		return false
	default:
		panic(fmt.Sprintf("unexpected type %T. ref isNode", t))
	}
}
func detect[T any](node sqlparser.SQLNode) T {
	var defautT T
	switch t := node.(type) {
	case T:
		return node.(T)
	case *sqlparser.AliasedExpr:
		return detect[T](t.Expr)
	case *sqlparser.ColName:
		return defautT
	case *sqlparser.BinaryExpr:
		return defautT
	default:
		panic(fmt.Sprintf("unexpected type %T. ref detect", t))
	}
}
func extractNodes[T any](selectExprs sqlparser.SelectExprs) []sqlparser.SQLNode {
	nodes := []sqlparser.SQLNode{}
	for _, expr := range selectExprs {
		if isNode[T](expr) {
			nodes = append(nodes, expr)
		}
	}
	return nodes
}

func (s *smarty) ToText(node sqlparser.SQLNode) string {
	if node == nil {
		return "<nil>"
	}
	buf := sqlparser.TrackedBuffer{
		Buffer: bytes.NewBuffer(nil),
	}
	node.Format(&buf)
	return buf.String()
}

var smartier = &smarty{}
