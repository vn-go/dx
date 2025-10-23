package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type sorting struct{}

func (s *sorting) resolveOrderBy(exprs sqlparser.OrderBy, injector *injector, reverse dictionaryFields) (*compilerResult, error) {
	items := []string{}
	for _, x := range exprs {
		switch fx := x.Expr.(type) {
		case *sqlparser.ColName:
			return exp.resolve(fx, injector, CMP_ORDER_BY, reverse)

		default:
			panic(fmt.Sprintf("not implemented: %T. See sorting.resolveOrderBy", fx))
		}
	}
	return &compilerResult{
		Content: strings.Join(items, ", "),
	}, nil
}

var sort = &sorting{}
