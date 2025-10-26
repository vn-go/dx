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
			tableName := strings.Trim(s.ToText(a.Expr), "`")
			//ret[strings.ToLower(tableName)] = strings.ToLower(tableName)
			if !a.As.IsEmpty() {
				ret[strings.ToLower(strings.ToLower(a.As.String()))] = strings.ToLower(tableName)
				//ret[strings.ToLower(tableName)] = strings.ToLower(strings.ToLower(a.As.String()))
			} else {
				ret[strings.ToLower(tableName)] = strings.ToLower(tableName)
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
		if t.Qualifier.IsEmpty() {
			// function name is table name or dataset name
			if _, ok := keywordFuncMap[strings.ToLower(t.Name.String())]; !ok {
				return []string{
					t.Name.String(),
				}
			}

		} else if strings.ToLower(t.Qualifier.String()) == "dataset" {
			// if function is the same of keyword nedd dataset qualifier
			return []string{
				t.Name.String(),
			}
		}

		return []string{}

	case sqlparser.SelectExprs:
		r := []string{}
		for _, expr := range t {
			r = append(r, s.extractTables(expr, visitedTable)...)
		}
		return r
	case *sqlparser.SQLVal:
		return []string{}
	case *sqlparser.AndExpr:
		r := s.extractTables(t.Left, visitedTable)
		r = append(r, s.extractTables(t.Right, visitedTable)...)
		return r
	case *sqlparser.OrExpr:
		r := s.extractTables(t.Left, visitedTable)
		r = append(r, s.extractTables(t.Right, visitedTable)...)
		return r
	case *sqlparser.BinaryExpr:
		r := s.extractTables(t.Left, visitedTable)
		r = append(r, s.extractTables(t.Right, visitedTable)...)
		return r

	default:
		panic(fmt.Sprintf("unknown type %T, smarty.extractTables", t))
	}
}

// smarty.convertToJoinTableExpr.go
func (s *smarty) convertToJoinTableExpr(comparisionNodes []joinCondition, tableAliasNodes []sqlparser.SQLNode, subSetInfoList map[string]subsetInfo) string {

	mapTableAlias := s.getMapTableAlias(tableAliasNodes)
	if len(comparisionNodes) == 0 {

		items := []string{}
		for k, v := range mapTableAlias {
			if n, ok := subSetInfoList[strings.ToLower(v)]; ok {
				items = append(items, "("+n.query+") "+sqlparser.Backtick(k))
			} else {
				items = append(items, v+" "+sqlparser.Backtick(k))
			}

		}
		return strings.Join(items, ", ")
	}

	strOn := s.ToText(comparisionNodes[0].node)
	tables := s.extractTables(comparisionNodes[0].node, map[string]bool{})
	strLeft := tables[0]
	tableHasUsed := map[string]bool{}
	tableHasUsed[strLeft] = true
	if aliasLeft, ok := mapTableAlias[strLeft]; ok {
		if n, ok := subSetInfoList[strings.ToLower(aliasLeft)]; ok {
			strLeft = "(" + n.query + ") " + sqlparser.Backtick(aliasLeft)
		} else {
			strLeft = aliasLeft + " " + sqlparser.Backtick(strLeft)
		}

		tableHasUsed[aliasLeft] = true
	} else if subset, ok := subSetInfoList[strings.ToLower(strLeft)]; ok {
		strLeft = "(" + subset.query + ") " + sqlparser.Backtick(subset.alias)
	}
	strRight := tables[1]
	tableHasUsed[strRight] = true
	if aliasRight, ok := mapTableAlias[strRight]; ok {
		if n, ok := subSetInfoList[strings.ToLower(aliasRight)]; ok {
			strRight = "(" + n.query + ") " + sqlparser.Backtick(aliasRight)
		} else {
			strRight = aliasRight + " " + sqlparser.Backtick(strRight)
		}

		tableHasUsed[aliasRight] = true

	} else if subset, ok := subSetInfoList[strings.ToLower(strRight)]; ok {
		strRight = "(" + subset.query + ") " + sqlparser.Backtick(subset.alias)
	}

	strJoin := strLeft + " " + comparisionNodes[0].joinType + " JOIN " + strRight + " ON " + strOn
	for i := 1; i < len(comparisionNodes); i++ {
		tables := s.extractTables(comparisionNodes[i].node, map[string]bool{})
		nextTable := tables[1]
		for _, table := range tables {
			if !tableHasUsed[table] {
				nextTable = table
				tableHasUsed[nextTable] = true
			}

		}

		if aliasTable, ok := mapTableAlias[nextTable]; ok {
			if n, ok := subSetInfoList[strings.ToLower(aliasTable)]; ok {
				nextTable = "(" + n.query + ") " + sqlparser.Backtick(aliasTable)
			} else {
				nextTable = aliasTable + " " + sqlparser.Backtick(nextTable)
			}
			tableHasUsed[aliasTable] = true
		} else if subset, ok := subSetInfoList[strings.ToLower(nextTable)]; ok {
			nextTable = "(" + subset.query + ") " + sqlparser.Backtick(subset.alias)
		}

		joinNext := s.ToText(comparisionNodes[i].node)
		strJoin += " " + comparisionNodes[i].joinType + " JOIN " + nextTable + " ON " + joinNext
	}

	return strJoin

}
