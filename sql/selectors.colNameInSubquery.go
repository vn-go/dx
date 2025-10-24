package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// selectors.colNameInSubquery.go
func (s selectors) colNameInSubquery(t *sqlparser.ColName, injector *injector, qrEntity subqueryEntity) (*compilerResult, error) {
	key := strings.ToLower(fmt.Sprintf("%s.%s", t.Qualifier.Name.String(), t.Name.String()))
	if _, ok := qrEntity.fields[key]; ok {
		ret := compilerResult{
			Content:         injector.dialect.Quote(t.Qualifier.Name.String(), t.Name.String()),
			OriginalContent: fmt.Sprintf("%s.%s", t.Qualifier.Name.String(), t.Name.String()),
			selectedExprsReverse: dictionaryFields{
				key: &dictionaryField{
					Expr: injector.dialect.Quote(t.Qualifier.Name.String(), t.Name.String()),
				},
			},
		}
		return &ret, nil
	} else {
		return nil, newCompilerError(ERR_FIELD_NOT_FOUND, "Column '%s' not found in sub dataset '%s", t.Name.String(), t.Qualifier.Name.String())
	}
}
