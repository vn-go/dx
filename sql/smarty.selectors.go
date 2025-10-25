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
func (s *smarty) selectors(selectStm *sqlparser.Select) string {
	items := []string{}
	nodes := s.extractSelectNodes(selectStm)
	for _, node := range nodes {
		items = append(items, s.ToText(node))
	}
	if len(items) == 0 {
		return "*"
	}
	return strings.Join(items, ", ")
}

var keywordFuncMap = map[string]bool{
	"from":   true,
	"where":  true,
	"order":  true,
	"limit":  true,
	"offset": true,
	"group":  true,
}

func (s *smarty) extractSelectNodes(selectStm *sqlparser.Select) []sqlparser.SQLNode {
	nodes := []sqlparser.SQLNode{}
	for _, expr := range selectStm.SelectExprs {
		switch n := expr.(type) {
		case *sqlparser.AliasedExpr:
			if s.isSelecteNode(n.Expr) {
				nodes = append(nodes, n)
			}
		default:
			panic(fmt.Sprintf("unknown SelectExpr type: %T. ref smarty.extractSelectNodes", expr))
		}
	}
	return nodes
}

func (s *smarty) isSelecteNode(expr sqlparser.SQLNode) bool {
	switch expr := expr.(type) {
	case *sqlparser.FuncExpr:
		return !keywordFuncMap[string(expr.Name.Lowered())]
	default:
		return true
	}
}
