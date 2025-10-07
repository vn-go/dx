package compiler

import (
	"fmt"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type tabelExtractorTypes struct {
}

var tableExtractor = &tabelExtractorTypes{}

type getTablesFromSqlStruct struct {
	tables     []string
	isSubQuery bool
}

func (t *tabelExtractorTypes) getTablesFromSql(sql string, node sqlparser.SQLNode) (*getTablesFromSqlStruct, error) {
	return internal.OnceCall("tabelExtractorTypes/getTablesFromSql/"+sql, func() (*getTablesFromSqlStruct, error) {
		isSubQuery := false
		mapTable := make(map[string]bool)
		x := t.getTables(node, mapTable)
		return &getTablesFromSqlStruct{
			tables:     x.tables,
			isSubQuery: isSubQuery,
		}, nil

	})

}
func (t *tabelExtractorTypes) getTables(node sqlparser.SQLNode, visited map[string]bool) *getTablesFromSqlStruct {
	//sqlparser.TableExprs
	if node == nil {
		return nil
	}
	ret := []string{}
	if sqlSelect, ok := node.(*sqlparser.Select); ok {
		next := t.getTables(sqlSelect.From, visited)
		if next.isSubQuery {
			return next
		}
		ret = append(ret, next.tables...)
		next = t.getTables(sqlSelect.GroupBy, visited)
		ret = append(ret, next.tables...)
		if sqlSelect.Where != nil {
			next = t.getTables(sqlSelect.Where, visited)
			ret = append(ret, next.tables...)
		}
		if sqlSelect.Having != nil {
			next = t.getTables(sqlSelect.Having, visited)

			ret = append(ret, next.tables...)
		}
		next = t.getTables(sqlSelect.SelectExprs, visited)
		ret = append(ret, next.tables...)
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if tableExprs, ok := node.(sqlparser.TableExprs); ok {
		for _, n := range tableExprs {
			nextTbl := t.getTables(n, visited)
			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if joinTableExpr, ok := node.(*sqlparser.JoinTableExpr); ok {

		nextTbl := t.getTables(joinTableExpr.LeftExpr, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		nextTbl = t.getTables(joinTableExpr.RightExpr, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		nextTbl = t.getTables(joinTableExpr.Condition, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if aliasedTableExpr, ok := node.(*sqlparser.AliasedTableExpr); ok {
		if aliasedTableExpr.As.IsEmpty() {
			nextTbl := t.getTables(aliasedTableExpr.Expr, visited)
			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}
			return &getTablesFromSqlStruct{
				tables: ret,
			}
		} else {

			nextTbl := t.getTables(aliasedTableExpr.Expr, visited)
			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}
			return &getTablesFromSqlStruct{
				tables: ret,
			}

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
		return &getTablesFromSqlStruct{
			tables: ret,
		}
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
			return &getTablesFromSqlStruct{
				tables: ret,
			}
		}
	}
	//sqlparser.JoinCondition
	if joinCondition, ok := node.(sqlparser.JoinCondition); ok {
		nextTbl := t.getTables(joinCondition.On, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	//*sqlparser.ComparisonExpr
	if comparisonExpr, ok := node.(*sqlparser.ComparisonExpr); ok {

		nextTbl := t.getTables(comparisonExpr.Left, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}

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
			return &getTablesFromSqlStruct{
				tables: ret,
			}
		}
	}
	if selects, ok := node.(sqlparser.SelectExprs); ok {
		for _, x := range selects {
			nextTbl := t.getTables(x, visited)
			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}
			return &getTablesFromSqlStruct{
				tables: ret,
			}
		}

	}
	if x, ok := node.(*sqlparser.AliasedExpr); ok {
		nextTbl := t.getTables(x.Expr, visited)

		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if x, ok := node.(sqlparser.GroupBy); ok {
		for _, f := range x {
			nextTbl := t.getTables(f, visited)

			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}

		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if x, ok := node.(*sqlparser.Where); ok {
		if x == nil {
			return nil
		}
		nextTbl := t.getTables(x.Expr, visited)

		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if x, ok := node.(*sqlparser.JoinTableExpr); ok {
		nextTbl := t.getTables(x.LeftExpr, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		nextTbl = t.getTables(x.Condition, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		nextTbl = t.getTables(x.RightExpr, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}

	}
	if x, ok := node.(*sqlparser.SelectExprs); ok {
		for _, f := range *x {
			nextTbl := t.getTables(f, visited)
			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}

	}
	//*sqlparser.BinaryExpr
	if x, ok := node.(*sqlparser.BinaryExpr); ok {
		nextTbl := t.getTables(x.Left, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		nextTbl = t.getTables(x.Right, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}

		return &getTablesFromSqlStruct{
			tables: ret,
		}

	}
	if _, ok := node.(*sqlparser.SQLVal); ok {
		return nil
	}
	if fn, ok := node.(*sqlparser.FuncExpr); ok {
		for _, f := range fn.Exprs {
			nextTbl := t.getTables(f, visited)
			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if s, ok := node.(*sqlparser.StarExpr); ok {
		if s.TableName.IsEmpty() {
			return nil
		}
		ret = append(ret, s.TableName.Name.String())
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if s, ok := node.(*sqlparser.AndExpr); ok {

		nextTbl := t.getTables(s.Left, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		nextTbl = t.getTables(s.Right, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}

		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if s, ok := node.(*sqlparser.OrExpr); ok {
		nextTbl := t.getTables(s.Left, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		nextTbl = t.getTables(s.Right, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if s, ok := node.(*sqlparser.NotExpr); ok {
		nextTbl := t.getTables(s.Expr, visited)

		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if s, ok := node.(*sqlparser.IsExpr); ok {
		nextTbl := t.getTables(s.Expr, visited)
		if nextTbl != nil {
			ret = append(ret, nextTbl.tables...)
		}
		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if _, ok := node.(*sqlparser.Subquery); ok {
		// nextTbl := t.getTables(s.Select, visited)

		// *isSubQuey = true //<--set bang tru o day, nhung retrun sang buoc ke tiep no la false
		// return append(ret, next...)
		return &getTablesFromSqlStruct{
			isSubQuery: true,
		}
	}
	if _, ok := node.(*sqlparser.Union); ok {
		return nil
	}
	if s, ok := node.(*sqlparser.UpdateExprs); ok {
		for _, x := range *s {
			nextTbl := t.getTables(x, visited)

			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}

		}

		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}
	if s, ok := node.(sqlparser.UpdateExprs); ok {
		for _, x := range s {
			nextTbl := t.getTables(x, visited)

			if nextTbl != nil {
				ret = append(ret, nextTbl.tables...)
			}

		}

		return &getTablesFromSqlStruct{
			tables: ret,
		}
	}

	if s, ok := node.(*sqlparser.ParenExpr); ok {
		return t.getTables(s.Expr, visited)
	}
	//sqlparser.Expr
	panic(fmt.Sprintf("not implement node type %s ,%s", node, `compiler\tabelExtractorTypes.go`))

}
