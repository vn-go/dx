package from

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/ds/errors"

	"github.com/vn-go/dx/sqlparser"
)

func (f *fromClauseType) buildFromClauseInfoByComparisonExpr(expr *sqlparser.ComparisonExpr, injectInfo *InjectInfo) (*fromClauseInfo, error) {
	left, err := f.fromClauseTypeByExpr(expr.Left, injectInfo)
	if err != nil {
		return nil, err
	}
	right, err := f.fromClauseTypeByExpr(expr.Right, injectInfo)
	if err != nil {
		return nil, err
	}

	return &fromClauseInfo{
		left:     left.table,
		right:    right.table,
		on:       fmt.Sprintf("%s %s %s", left.content, expr.Operator, right.content),
		joinType: "inner",
		next:     nil,
		top:      nil,
	}, nil

}

type buildFromClauseInfoByExprResult struct {
	content         string
	originalContent string
	leftTable       string
	rightTable      string
}
type fromClauseTypeResult struct {
	content string
	table   string
}

func (f *fromClauseType) fromClauseTypeByExpr(expr sqlparser.Expr, injectInfo *InjectInfo) (*fromClauseTypeResult, error) {
	switch expr := expr.(type) {
	case *sqlparser.ColName:

		return f.fromClauseTypeColName(expr, injectInfo)

	}

	panic(fmt.Sprintf("not implemented yet: %T. see fromClauseType.buildFromClauseInfoByExpr", expr))
}

func (f *fromClauseType) fromClauseTypeColName(expr *sqlparser.ColName, injectInfo *InjectInfo) (*fromClauseTypeResult, error) {
	if expr.Qualifier.IsEmpty() {
		return nil, errors.NewParseError("'%s' requires dataset name", expr.Name.String())
	}
	tableName := expr.Qualifier.Name.String()
	source := ""
	dbTableNameToFindAlias := ""

	if dbTableName, ok := injectInfo.Dict.AliasMap[strings.ToLower(tableName)]; ok {
		dbTableNameToFindAlias = dbTableName
		if dbTableAlias, ok := injectInfo.Dict.TableAlias[dbTableName]; ok {

			source = injectInfo.Dialect.Quote(dbTableName) + " " + injectInfo.Dialect.Quote(dbTableAlias)

		} else {
			source = injectInfo.Dialect.Quote(dbTableName)
		}
	} else {
		injectInfo.Dict.BuildByAliasTableName(injectInfo.Dialect, "", tableName)

		if dbTableName, ok := injectInfo.Dict.AliasMap[strings.ToLower(tableName)]; ok {
			dbTableNameToFindAlias = dbTableName
			source = injectInfo.Dialect.Quote(dbTableName)
		} else {
			return nil, errors.NewParseError("dataset '%s' is not found", tableName)
		}

	}
	key := strings.ToLower(fmt.Sprintf("%s.%s", tableName, expr.Name.String()))
	if tableAlias, ok := injectInfo.Dict.TableAlias[strings.ToLower(dbTableNameToFindAlias)]; ok && tableAlias != "" {
		//when end user use tablename to refer to a table, we should use alias instead of original table name in the query
		key = strings.ToLower(fmt.Sprintf("%s.%s", tableAlias, expr.Name.String()))
	}
	if colName, ok := injectInfo.Dict.FieldMap[key]; ok {
		return &fromClauseTypeResult{
			content: colName.Content,
			table:   source,
		}, nil
	} else {
		return nil, errors.NewParseError("field '%s' is not found in dataset '%s'", expr.Name.String(), tableName)
	}

}
