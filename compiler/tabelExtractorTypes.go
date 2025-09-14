package compiler

import (
	"fmt"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type tabelExtractorTypes struct {
}

var tabelExtractor = &tabelExtractorTypes{}

func (t *tabelExtractorTypes) getTablesFromSql(sql string, node sqlparser.SQLNode) []string {
	ret, _ := internal.OnceCall("tabelExtractorTypes/getTablesFromSql/"+sql, func() ([]string, error) {
		x := t.getTables(node, make(map[string]bool))
		return x, nil
	})
	return ret
}
func (t *tabelExtractorTypes) getTables(node sqlparser.SQLNode, visited map[string]bool) []string {
	//sqlparser.TableExprs
	if node == nil {
		return []string{}
	}
	ret := []string{}
	if sqlSelect, ok := node.(*sqlparser.Select); ok {
		next := t.getTables(sqlSelect.From, visited)
		ret = append(ret, next...)
		next = t.getTables(sqlSelect.GroupBy, visited)
		ret = append(ret, next...)
		next = t.getTables(sqlSelect.Where, visited)
		ret = append(ret, next...)
		next = t.getTables(sqlSelect.Having, visited)
		ret = append(ret, next...)
		next = t.getTables(sqlSelect.SelectExprs, visited)
		ret = append(ret, next...)
		return ret
	}
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
		nextTbl = t.getTables(joinTableExpr.Condition, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
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

		ret = append(ret, next...)
		return ret
	}
	if x, ok := node.(sqlparser.GroupBy); ok {
		for _, f := range x {
			ret = append(ret, t.getTables(f, visited)...)
		}
		return ret
	}
	if x, ok := node.(*sqlparser.Where); ok {
		if x == nil {
			return []string{}
		}
		ret = append(ret, t.getTables(x.Expr, visited)...)
		return ret
	}
	if x, ok := node.(*sqlparser.JoinTableExpr); ok {
		ret = append(ret, t.getTables(x.LeftExpr, visited)...)
		ret = append(ret, t.getTables(x.Condition, visited)...)
		ret = append(ret, t.getTables(x.RightExpr, visited)...)
		return ret
	}
	if x, ok := node.(*sqlparser.SelectExprs); ok {
		for _, f := range *x {
			ret = append(ret, t.getTables(f, visited)...)
		}
		return ret

	}
	//*sqlparser.BinaryExpr
	if x, ok := node.(*sqlparser.BinaryExpr); ok {
		ret = append(ret, t.getTables(x.Left, visited)...)
		ret = append(ret, t.getTables(x.Right, visited)...)
		return ret

	}
	if _, ok := node.(*sqlparser.SQLVal); ok {
		return []string{}
	}
	if fn, ok := node.(*sqlparser.FuncExpr); ok {
		for _, f := range fn.Exprs {
			ret = append(ret, t.getTables(f, visited)...)
		}
		return ret
	}
	if s, ok := node.(*sqlparser.StarExpr); ok {
		if s.TableName.IsEmpty() {
			return []string{}
		}
		ret = append(ret, s.TableName.Name.String())
		return ret
	}
	if s, ok := node.(*sqlparser.AndExpr); ok {
		next := t.getTables(s.Left, visited)
		ret = append(ret, next...)
		next = t.getTables(s.Right, visited)
		ret = append(ret, next...)
		return ret
	}
	if s, ok := node.(*sqlparser.OrExpr); ok {
		next := t.getTables(s.Left, visited)
		ret = append(ret, next...)
		next = t.getTables(s.Right, visited)
		ret = append(ret, next...)
		return ret
	}
	if s, ok := node.(*sqlparser.NotExpr); ok {
		next := t.getTables(s.Expr, visited)
		ret = append(ret, next...)

		return ret
	}
	if s, ok := node.(*sqlparser.IsExpr); ok {
		next := t.getTables(s.Expr, visited)
		return append(ret, next...)
	}
	if s, ok := node.(*sqlparser.Subquery); ok {
		next := t.getTables(s.Select, visited)
		return append(ret, next...)
	}
	if s, ok := node.(*sqlparser.Union); ok {
		next := t.getTables(s.Left, visited)
		ret = append(ret, next...)
		next = t.getTables(s.Right, visited)
		ret = append(ret, next...)
		return ret
	}
	if s, ok := node.(*sqlparser.UpdateExprs); ok {
		for _, x := range *s {
			next := t.getTables(x, visited)

			ret = append(ret, next...)
		}

		return ret
	}
	if s, ok := node.(sqlparser.UpdateExprs); ok {
		for _, x := range s {
			next := t.getTables(x, visited)

			ret = append(ret, next...)
		}

		return ret
	}
	if s, ok := node.(*sqlparser.UpdateExpr); ok {
		next := t.getTables(s.Expr, visited)
		if !s.Name.Qualifier.IsEmpty() {
			next = append(next, s.Name.Qualifier.Name.String())
		}

		ret = append(ret, next...)
		return ret
	}
	//sqlparser.Expr
	panic(fmt.Sprintf("not implement ,%s", `compiler\tabelExtractorTypes.go`))

}
