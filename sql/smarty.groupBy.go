package sql

import (
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// smarty.groupBy.go
func (s *smarty) groupBy(selectStm *sqlparser.Select) string {
	items := []string{}
	for _, expr := range selectStm.SelectExprs {
		if x := detect[*sqlparser.FuncExpr](expr); x != nil {
			if x.Name.Lowered() == "group" {

				for _, expr := range x.Exprs {

					items = append(items, s.ToText(expr))
				}
			}
		}
	}
	return strings.Join(items, ", ")
}
