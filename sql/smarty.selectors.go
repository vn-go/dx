package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

func (s *smarty) where(selectStm *sqlparser.Select) string {
	for _, expr := range selectStm.SelectExprs {
		if fn := detect[*sqlparser.FuncExpr](expr); fn != nil {
			if fn.Name.Lowered() == "where" {
				return s.ToText(fn.Exprs[0])
			}
		}
	}
	return ""
}

// smarty.selectors.go
func (s *smarty) selectors(selectStm *sqlparser.Select, fieldAliasMap map[string]string) (string, error) {
	items := []string{}
	nodes := s.extractSelectNodes(selectStm)
	for _, node := range nodes {
		if fn := detect[*sqlparser.FuncExpr](node); fn != nil {
			if fn.Name.Lowered() == "crosstab" {
				return crossTabs.resolve(node.(*sqlparser.AliasedExpr), fn, fieldAliasMap)
			}
			if !fn.Qualifier.IsEmpty() {
				if strings.ToLower(fn.Qualifier.String()) == "dataset" {
					for _, x := range fn.Exprs {
						nodeWithTable := tableApplier.resolve(x, fn.Name.String())
						strExpr := s.ToText(nodeWithTable)
						items = append(items, strExpr)
					}
					continue
				}
			} else if aliasNode, ok := node.(*sqlparser.AliasedExpr); ok {
				if aliasNode.As.IsEmpty() { // means function name is table name
					if _, ok := keywordFuncMap[fn.Name.Lowered()]; !ok {
						for _, x := range fn.Exprs {
							nodeWithTable := tableApplier.resolve(x, fn.Name.String())
							strExpr := s.ToText(nodeWithTable)

							items = append(items, strExpr)
						}
						continue
					}
				}

			}
		}
		if fx, ok := node.(*sqlparser.AliasedExpr); ok {
			if !fx.As.IsEmpty() {
				fieldAliasMap[string(fx.As.Lowered())] = s.ToText(fx.Expr)
			}
		}
		items = append(items, s.ToText(node))
	}
	if len(items) == 0 {
		return "*", nil
	}
	return strings.Join(items, ", "), nil
}

func (s *smarty) extractSelectNodes(selectStm *sqlparser.Select) []sqlparser.SQLNode {
	nodes := []sqlparser.SQLNode{}
	for _, expr := range selectStm.SelectExprs {
		switch n := expr.(type) {
		case *sqlparser.AliasedExpr:

			if s.isSelecteNode(n.Expr) {
				nodes = append(nodes, n)
			}
		// case *sqlparser.FuncExpr:

		default:
			panic(fmt.Sprintf("unknown SelectExpr type: %T. ref smarty.extractSelectNodes", expr))
		}
	}
	return nodes
	/*
		if fn := detect[*sqlparser.FuncExpr](node); fn != nil {

			}
	*/
}

func (s *smarty) isSelecteNode(expr sqlparser.SQLNode) bool {
	switch expr := expr.(type) {
	case *sqlparser.FuncExpr:
		return !keywordFuncMap[string(expr.Name.Lowered())]
	default:
		return true
	}
}
