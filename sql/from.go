package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type from struct {
}

var froms = &from{}

func (f *from) resolve(expr sqlparser.TableExprs, injector *injector) (*compilerResult, error) {
	for i, x := range expr {
		switch t := x.(type) {
		case *sqlparser.AliasedTableExpr:
			alias := fmt.Sprintf("T%d", i+1)
			if !t.As.IsEmpty() {
				alias = strings.ToLower(t.As.String())
			}
			return f.AliasedTableExpr(t, alias, injector)
		default:
			panic(fmt.Sprintf("unsupported table expression type: %T. see from.resolve, %s", t, `sql\from.go`))
		}
	}
	return nil, nil

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

	// case sqlparser.SimpleTableExpr:
	//

	default:
		panic(fmt.Sprintf("unsupported table expression type: %T. see from.AliasedTableExpr, %s", t, `sql\from.go`))
	}
	panic(fmt.Sprintf("not implemented. see from.AliasedTableExpr, %s", `sql\from.go`))
}
