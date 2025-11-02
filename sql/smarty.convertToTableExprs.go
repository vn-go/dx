package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// smarty.convertToTableExprs.go
func (s *smarty) convertToTableExprs(exprs *sqlparser.Select, subSetInfoList map[string]subsetInfo) string {
	visited := make(map[string]bool)
	items := []string{}
	for _, expr := range exprs.SelectExprs {
		sourceNames := s.extractTables(expr, visited)
		for _, sourceName := range sourceNames {
			if subsets, ok := subSetInfoList[strings.ToLower(sourceName)]; ok {
				querySource := fmt.Sprintf("(%s)  %s", subsets.query, subsets.alias)
				items = append(items, querySource)
				continue
			}
			items = append(items, sourceName)
		}
	}

	return strings.Join(items, ", ")

}
