package sql

import (
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// smarty.groupBy.go
func (s *smarty) groupBy(selectStm *sqlparser.Select, fieldAliasMap map[string]string) string {
	items := []string{}
	for _, expr := range selectStm.SelectExprs {
		if x := detect[*sqlparser.FuncExpr](expr); x != nil {
			if x.Name.Lowered() == "group" {

				for _, expr := range x.Exprs {
					if a, ok := expr.(*sqlparser.AliasedExpr); ok {
						strExpr := s.ToText(a.Expr)
						if fx, ok := fieldAliasMap[strings.Trim(strExpr, "`")]; ok {
							items = append(items, fx)
						} else {
							items = append(items, strExpr)
							fieldAliasMap[strings.Trim(a.As.Lowered(), "`")] = strExpr
						}

					} else {
						items = append(items, s.ToText(expr))
					}

				}
			}
		}
	}
	return strings.Join(items, ", ")
}
