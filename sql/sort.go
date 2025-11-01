package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type sorting struct{}

func (s *sorting) resolveOrderBy(exprs sqlparser.OrderBy, injector *injector, reverse *dictionaryFields) (*compilerResult, error) {
	items := []string{}
	for _, x := range exprs {
		switch fx := x.Expr.(type) {
		case *sqlparser.ColName:
			r, err := exp.resolve(fx, injector, CMP_ORDER_BY, reverse)
			if err != nil {
				return nil, err
			}
			items = append(items, r.Content+" "+x.Direction)
		case *sqlparser.FuncExpr:
			r, err := exp.resolve(fx, injector, CMP_ORDER_BY, reverse)
			if err != nil {
				return nil, err
			}
			items = append(items, r.Content+" "+x.Direction)

		default:
			panic(fmt.Sprintf("not implemented: %T. See sorting.resolveOrderBy", fx))
		}
	}
	return &compilerResult{
		Content: strings.Join(items, ", "),
	}, nil
}

var sort = &sorting{}
