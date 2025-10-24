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
		case *sqlparser.JoinTableExpr:

			return f.joinTableExpr(t, injector, &joinTableExprInjector{})
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
			return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "dataset from by another stetement must have alias")
		}
		alias = expr.As.String()
		return f.subquery(t, alias, injector)

	default:
		panic(fmt.Sprintf("unsupported table expression type: %T. see from.AliasedTableExpr, %s", t, `sql\from.go`))
	}

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
