package sql

import (
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// smarty.sort.go
func (s *smarty) sort(selectStm *sqlparser.Select, fieldAliasMap map[string]string) string {
	items := []string{}
	for _, expr := range selectStm.SelectExprs {
		if fn := detect[*sqlparser.FuncExpr](expr); fn != nil {
			if fn.Name.Lowered() == "sort" {
				for _, x := range fn.Exprs {
					switch x := x.(type) {
					case *sqlparser.AliasedExpr:
						exprSort := s.ToText(x.Expr)
						if fx, ok := fieldAliasMap[strings.ToLower(strings.Trim(exprSort, "`"))]; ok {
							exprSort = fx
						}
						sortType := "asc"
						if !x.As.IsEmpty() {
							sortType = strings.Trim(s.ToText(x.As), "`")

						}

						items = append(items, exprSort+" "+sortType)
					default:
						panic("not support sort type")

					}

				}

			}
		}
	}
	return strings.Join(items, ", ")
}
