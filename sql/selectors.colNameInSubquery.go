package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// selectors.colNameInSubquery.go
func (s selectors) colNameInSubquery(t *sqlparser.ColName, injector *injector, cmpTyp CMP_TYP, selectedExprsReverse dictionaryFields, qrEntity subqueryEntity) (*compilerResult, error) {
	key := strings.ToLower(fmt.Sprintf("%s.%s", t.Qualifier.Name.String(), t.Name.String()))
	return &compilerResult{
		Content:         injector.dialect.Quote(t.Qualifier.Name.String(), t.Name.String()),
		OriginalContent: fmt.Sprintf("%s.%s", t.Qualifier.Name.String(), t.Name.String()),
		selectedExprsReverse: dictionaryFields{
			key: &dictionaryField{
				Expr: injector.dialect.Quote(t.Qualifier.Name.String(), t.Name.String()),
			},
		},
	}, nil
}
