package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// from.joinTableExpr.go
func (f *from) joinTableExpr(t *sqlparser.JoinTableExpr, injector *injector, joinInjector *joinTableExprInjector) (*compilerResult, error) {

	left, err := f.tableExpr(t.LeftExpr, injector, joinInjector)
	if err != nil {
		return nil, err
	}

	right, err := f.tableExpr(t.RightExpr, injector, joinInjector)
	if err != nil {
		return nil, err
	}
	selectedExprsReverse := dictionaryFields{}
	condition, err := exp.resolve(t.Condition.On, injector, CMP_JOIN, &selectedExprsReverse)
	if err != nil {
		return nil, err
	}
	condition.Fields = condition.Fields.merge(left.Fields.merge(right.Fields))
	return &compilerResult{
		Content: fmt.Sprintf("%s %s  %s ON %s", left.Content, t.Join, right.Content, condition.Content),
		Fields:  condition.Fields,
	}, nil
}

func (f *from) tableExpr(expr sqlparser.TableExpr, injector *injector, joinInjector *joinTableExprInjector) (*compilerResult, error) {
	switch x := expr.(type) {
	case *sqlparser.AliasedTableExpr:

		tableAlias := fmt.Sprintf("T%d", joinInjector.index+1)
		if !x.As.IsEmpty() {
			tableAlias = strings.ToLower(x.As.String())
		} else {
			joinInjector.index++
		}
		return f.aliasedTableExpr(x, injector, tableAlias)
	case *sqlparser.JoinTableExpr:
		ret, err := f.joinTableExpr(x, injector, joinInjector)
		if err != nil {
			return nil, traceCompilerError(err, contents.toText(x))
		}
		return ret, nil
	default:
		panic(fmt.Sprintf("not implemented %T, see from.tableExpr", x))
	}

}

func (f *from) aliasedTableExpr(expr *sqlparser.AliasedTableExpr, injector *injector, tableAlias string) (*compilerResult, error) {

	switch x := expr.Expr.(type) {
	case sqlparser.TableName:
		return f.simpleTableExpr(x, tableAlias, injector)
	case *sqlparser.Subquery:
		return froms.subquery(x, tableAlias, injector)

	default:
		panic(fmt.Sprintf("not implemented %T, see from.aliasedTableExpr", x))
	}

}

func (f *from) simpleTableExpr(expr sqlparser.TableName, alias string, injector *injector) (*compilerResult, error) {
	table := expr.Name.String()
	injector.dict.Build(alias, table, injector.dialect)
	dbTablename, ok := injector.dict.entities[strings.ToLower(table)]
	if !ok {
		return nil, newCompilerError(ERR_DATASET_NOT_FOUND, "dataset '%s' not found", table)
	}
	return &compilerResult{
		OriginalContent: table,
		Content:         injector.dialect.Quote(dbTablename.TableName) + " " + injector.dialect.Quote(alias),
	}, nil

}
