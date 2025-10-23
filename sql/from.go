package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type from struct {
}

var froms = &from{}

func (f *from) resolve(expr sqlparser.TableExprs, injector *injector) (*compilerResult, error) {
	ret := &compilerResult{}
	items := []string{}
	for i, x := range expr {
		switch t := x.(type) {
		case *sqlparser.AliasedTableExpr:
			alias := fmt.Sprintf("T%d", i+1)
			if !t.As.IsEmpty() {
				alias = strings.ToLower(t.As.String())
			}
			r, err := f.AliasedTableExpr(t, alias, injector)
			if err != nil {
				return nil, err
			}
			items = append(items, r.Content)
			ret.Fields = internal.UnionMap(ret.Fields, r.Fields)

		default:
			panic(fmt.Sprintf("unsupported table expression type: %T. see from.resolve, %s", t, `sql\from.go`))
		}
	}
	ret.Content = strings.Join(items, ", ")
	return ret, nil

}

func (f *from) AliasedTableExpr(expr *sqlparser.AliasedTableExpr, alias string, injector *injector) (*compilerResult, error) {
	switch t := expr.Expr.(type) {
	case sqlparser.TableName:
		buildResult, err := injector.dict.Build(alias, t.Name.String(), injector.dialect)
		if err != nil {
			return nil, err
		}

		return &compilerResult{
			Content: buildResult.backtick(injector.dialect),
		}, nil

	case *sqlparser.Subquery:
		if expr.As.IsEmpty() {
			return nil, newCompilerError("dataset from by another stetement must have alias")
		}
		alias = expr.As.String()
		backupDick := injector.dict

		defer func() {

			backupDick.fields = internal.UnionMap(backupDick.fields, injector.dict.fields)
			backupDick.tableAlias = internal.UnionMap(backupDick.tableAlias, injector.dict.tableAlias)
			injector.dict = backupDick

		}()

		injector.dict = newDictionary() // sub query need new dictionary for compiling
		ret, err := f.selectStatement(t.Select, injector)
		if err != nil {
			return nil, err
		}
		injector.dict.tableAlias[strings.ToLower(alias)] = alias
		for _, x := range ret.selectedExprs {
			key := strings.ToLower(fmt.Sprintf("%s.%s", alias, x.Alias))
			injector.dict.fields[key] = &dictionaryField{
				Expr:  injector.dialect.Quote(alias, x.Alias),
				Typ:   x.Typ,
				Alias: x.Alias,
			}
		}
		ret.Content = fmt.Sprintf("(%s) %s", ret.Content, injector.dialect.Quote(alias))
		return ret, nil

	default:
		panic(fmt.Sprintf("unsupported table expression type: %T. see from.AliasedTableExpr, %s", t, `sql\from.go`))
	}
	//panic(fmt.Sprintf("not implemented. see from.AliasedTableExpr, %s", `sql\from.go`))
}

func (f *from) selectStatement(sqlStm sqlparser.Statement, injector *injector) (*compilerResult, error) {
	switch expr := sqlStm.(type) {
	case *sqlparser.Select:
		return selector.selects(expr, injector)
	case *sqlparser.Union:
		return selector.union(expr, injector)

	default:
		panic(fmt.Sprintf("not support statement type: %T. see compiler.Resolve", sqlStm))

	}
}
