package sql

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type smarty struct {
}

var joinTypeFuncMap = map[string]string{
	"left":  "LEFT",
	"right": "RIGHT",
	"inner": "INNER",

	"outer": "OUTER",
	"full":  "`FULL OUTER`",
}

type joinCondition struct {
	node     sqlparser.SQLNode
	joinType string
}

func (s *smarty) from(selectStm *sqlparser.Select, subSetInfoList map[string]subsetInfo) string {

	dsSourceFunc := s.findFncExpr(selectStm, "from")
	if dsSourceFunc != nil {
		comparisonExprs := []joinCondition{}
		aliasedExpr := []sqlparser.SQLNode{}
		for _, expr := range dsSourceFunc.Exprs {
			if fnNode := detect[*sqlparser.FuncExpr](expr); fnNode != nil {
				if _, ok := joinTypeFuncMap[fnNode.Name.Lowered()]; ok {
					if fnNode.Name.Lowered() == "full" {
						// sqlparser can not parse full outer join, so we need to convert it to left outer join and right outer join
						fnNode.Qualifier = sqlparser.NewTableIdent(FIX_ERROR)
						fnNode.Name = sqlparser.NewColIdent(sqlparser.Backtick(FIX_ERROR_FULL_JOIN))
						comparisonExprs = append(comparisonExprs, joinCondition{
							node: fnNode,
						})
						continue
					}

				}
			}
			if isNode[*sqlparser.ComparisonExpr](expr) {
				comparisonExprs = append(comparisonExprs, joinCondition{
					node:     expr,
					joinType: "INNER",
				})
			} else if aliased, ok := expr.(*sqlparser.AliasedExpr); ok {
				aliasedExpr = append(aliasedExpr, aliased)
			} else {
				panic(fmt.Sprintf("unexpected type %T. ref smarty.from", expr))
			}
		}

		if len(comparisonExprs) > 0 || len(aliasedExpr) > 0 {
			return s.convertToJoinTableExpr(comparisonExprs, aliasedExpr, subSetInfoList)
		}
	} else {
		// if not found from function, then detect table name by function name
		// example:
		// dsl is "user(id, name),where(id=1)"
		return s.convertToTableExprs(selectStm, subSetInfoList)
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
	case *sqlparser.BinaryExpr:
		return false
	case *sqlparser.AndExpr:
		return false
	case *sqlparser.OrExpr:
		return false
	case *sqlparser.FuncExpr:
		return true
		// if t.Qualifier.IsEmpty() {
		// 	return true
		// } else {
		// 	return false
		// }

	default:
		panic(fmt.Sprintf("unexpected type %T. ref isNode", t))
	}
}
func find[T any](node sqlparser.SQLNode) (T, bool) {
	var defautT T
	switch t := node.(type) {
	case T:
		return node.(T), true
	case *sqlparser.AliasedExpr:
		return find[T](t.Expr)
	case *sqlparser.ColName:
		if reflect.TypeFor[T]() == reflect.TypeFor[*sqlparser.ColName]() {
			return node.(T), true
		}
		return defautT, false
	case *sqlparser.BinaryExpr:
		if ret, found := find[T](t.Left); found {
			return ret, found
		}
		return find[T](t.Right)
	case *sqlparser.ComparisonExpr:
		if ret, found := find[T](t.Left); found {
			return ret, found
		}
		return find[T](t.Right)
	case *sqlparser.FuncExpr:
		for _, expr := range t.Exprs {
			if ret, found := find[T](expr); found {
				return ret, found
			}
		}
		return defautT, false
	case *sqlparser.AndExpr:
		if ret, found := find[T](t.Left); found {
			return ret, found
		}
		return find[T](t.Right)
	case *sqlparser.OrExpr:
		if ret, found := find[T](t.Left); found {
			return ret, found
		}
		return find[T](t.Right)
	case *sqlparser.SQLVal:
		if reflect.TypeFor[T]() == reflect.TypeFor[*sqlparser.SQLVal]() {
			return node.(T), true
		}

		return defautT, false
	default:
		panic(fmt.Sprintf("unexpected type %T. ref detect", t))
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
	case *sqlparser.ComparisonExpr:
		return defautT
	case *sqlparser.FuncExpr:
		return defautT
	case *sqlparser.AndExpr:
		return defautT
	case *sqlparser.OrExpr:
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
