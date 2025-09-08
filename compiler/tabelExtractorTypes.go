package compiler

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

type tabelExtractorTypes struct {
}

var tabelExtractor = &tabelExtractorTypes{}

func (t *tabelExtractorTypes) getTables(node sqlparser.SQLNode, visited map[string]bool) []string {
	//sqlparser.TableExprs
	ret := []string{}
	if tableExprs, ok := node.(sqlparser.TableExprs); ok {
		for _, n := range tableExprs {
			nextTbl := t.getTables(n, visited)
			if len(nextTbl) > 0 {
				ret = append(ret, nextTbl...)
			}
		}
		return ret
	}
	if joinTableExpr, ok := node.(*sqlparser.JoinTableExpr); ok {

		nextTbl := t.getTables(joinTableExpr.LeftExpr, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		nextTbl = t.getTables(joinTableExpr.RightExpr, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		// nextTbl = t.getTables(joinTableExpr.Condition, visited)
		// if len(nextTbl) > 0 {
		// 	ret = append(ret, nextTbl...)
		// }
		return ret
	}
	if aliasedTableExpr, ok := node.(*sqlparser.AliasedTableExpr); ok {
		if aliasedTableExpr.As.IsEmpty() {
			nextTbl := t.getTables(aliasedTableExpr.Expr, visited)
			if len(nextTbl) > 0 {
				ret = append(ret, nextTbl...)
			}
			return ret
		} else {
			nextTbl := t.getTables(aliasedTableExpr.Expr, visited)
			if len(nextTbl) > 0 {
				for _, t := range nextTbl {
					ret = append(ret, fmt.Sprintf("%s\n%s", t, aliasedTableExpr.As.String()))
				}

			}
			return ret
		}

	}
	//sqlparser.TableIdent
	if tableIdent, ok := node.(sqlparser.TableIdent); ok {
		if tableIdent.IsEmpty() {
			return nil
		} else {
			if _, ok := visited[tableIdent.String()]; !ok {
				visited[tableIdent.String()] = true
				ret = append(ret, tableIdent.String())
			}

		}
		return ret
	}
	//sqlparser.TableName
	if tableName, ok := node.(sqlparser.TableName); ok {
		if tableName.IsEmpty() {
			return nil
		} else {
			if _, ok := visited[tableName.Name.String()]; !ok {
				visited[tableName.Name.String()] = true
				ret = append(ret, tableName.Name.String())
			}
			return ret
		}
	}
	//sqlparser.JoinCondition
	if joinCondition, ok := node.(sqlparser.JoinCondition); ok {
		nextTbl := t.getTables(joinCondition.On, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		return ret
	}
	//*sqlparser.ComparisonExpr
	if comparisonExpr, ok := node.(*sqlparser.ComparisonExpr); ok {

		nextTbl := t.getTables(comparisonExpr.Left, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		nextTbl = t.getTables(comparisonExpr.Right, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		return ret
	}
	//*sqlparser.ColName
	if colName, ok := node.(*sqlparser.ColName); ok {
		if colName.Qualifier.IsEmpty() {
			return nil
		} else {
			if _, ok := visited[colName.Qualifier.Name.String()]; !ok {
				visited[colName.Qualifier.Name.String()] = true
				ret = append(ret, colName.Qualifier.Name.String())
			}
			return ret
		}
	}
	if selects, ok := node.(sqlparser.SelectExprs); ok {
		for _, x := range selects {
			next := t.getTables(x, visited)
			if len(next) > 0 {
				ret = append(ret, next...)
			}
		}
		return ret
	}
	if x, ok := node.(*sqlparser.AliasedExpr); ok {
		next := t.getTables(x.Expr, visited)
		if !x.As.IsEmpty() {
			next = append(next, x.As.String())
		}
		ret = append(ret, next...)
		return ret
	}

	//sqlparser.Expr
	panic("not implement")

}
