package sql

import (
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// smarty.convertToTableExprs.go
func (s *smarty) convertToTableExprs(exprs *sqlparser.Select, subSetInfoList map[string]subsetInfo) string {
	visited := make(map[string]bool)
	items := []string{}
	for _, expr := range exprs.SelectExprs {
		items = append(items, s.extractTables(expr, visited)...)
	}

	return strings.Join(items, ", ")

}
