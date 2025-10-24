package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

func (s *smarty) getMapTableAlias(tableAliasNodes []sqlparser.SQLNode) map[string]string {
	ret := make(map[string]string)
	for _, node := range tableAliasNodes {
		if a, ok := node.(*sqlparser.AliasedExpr); ok {
			tableName := s.ToText(a.Expr)
			//ret[strings.ToLower(tableName)] = strings.ToLower(tableName)
			if !a.As.IsEmpty() {
				ret[strings.ToLower(strings.ToLower(a.As.String()))] = strings.ToLower(tableName)
				//ret[strings.ToLower(tableName)] = strings.ToLower(strings.ToLower(a.As.String()))
			}
		}
	}
	return ret
}
func (s *smarty) extractTables(node sqlparser.SQLNode, visitedTable map[string]bool) []string {
	switch t := node.(type) {
	case *sqlparser.AliasedExpr:
		return s.extractTables(t.Expr, visitedTable)
	case *sqlparser.ComparisonExpr:
		r := s.extractTables(t.Left, visitedTable)
		r = append(r, s.extractTables(t.Right, visitedTable)...)
		return r
	case *sqlparser.ColName:
		if !t.Qualifier.IsEmpty() {
			if !visitedTable[strings.ToLower(t.Qualifier.Name.String())] {
				visitedTable[strings.ToLower(t.Qualifier.Name.String())] = true
				return []string{strings.ToLower(t.Qualifier.Name.String())}
			} else {
				return []string{}
			}
		} else {
			return []string{}
		}
	case *sqlparser.FuncExpr:
		return s.extractTables(t.Exprs, visitedTable)
	case sqlparser.SelectExprs:
		r := []string{}
		for _, expr := range t {
			r = append(r, s.extractTables(expr, visitedTable)...)
		}
		return r
	case *sqlparser.SQLVal:
		return []string{}
	default:
		panic(fmt.Sprintf("unknown type %T, smarty.extractTables", t))
	}
}

// smarty.convertToJoinTableExpr.go
func (s *smarty) convertToJoinTableExpr(comparisionNodes []sqlparser.SQLNode, tableAliasNodes []sqlparser.SQLNode) string {

	mapTableAlias := s.getMapTableAlias(tableAliasNodes)
	if len(comparisionNodes) == 0 {
		items := []string{}
		for k, v := range mapTableAlias {
			items = append(items, v+" "+k)
		}
		return strings.Join(items, ", ")
	}

	strOn := s.ToText(comparisionNodes[0])
	tables := s.extractTables(comparisionNodes[0], map[string]bool{})
	strLeft := tables[0]
	tableHasUsed := map[string]bool{}
	tableHasUsed[strLeft] = true
	if aliasLeft, ok := mapTableAlias[strLeft]; ok {
		strLeft = aliasLeft + " " + strLeft
		tableHasUsed[aliasLeft] = true
	}
	strRight := tables[1]
	tableHasUsed[strRight] = true
	if aliasRight, ok := mapTableAlias[strRight]; ok {
		strRight = aliasRight + " " + strRight
		tableHasUsed[aliasRight] = true

	}

	strJoin := strLeft + " " + "INNER JOIN " + strRight + " ON " + strOn
	for i := 1; i < len(comparisionNodes); i++ {
		tables := s.extractTables(comparisionNodes[i], map[string]bool{})
		nextTable := tables[1]
		for _, table := range tables {
			if !tableHasUsed[table] {
				nextTable = table
				tableHasUsed[nextTable] = true
			}

		}

		if aliasTable, ok := mapTableAlias[nextTable]; ok {
			nextTable = aliasTable + " " + nextTable
			tableHasUsed[aliasTable] = true
		}

		joinNext := s.ToText(comparisionNodes[i])
		strJoin += " " + "INNER JOIN " + nextTable + " ON " + joinNext
	}

	return strJoin

}
