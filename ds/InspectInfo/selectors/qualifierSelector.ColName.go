package selectors

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/ds/errors"
	"github.com/vn-go/dx/sqlparser"
)

// qualifierSelector.ResolveColName.go
func (s *qualifierSelector) ColName(expr *sqlparser.ColName, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	key := strings.ToLower(fmt.Sprintf("%s.%s", s.Name, expr.Name.String()))

	if field, ok := injectInfo.Dict.FieldMap[key]; ok {
		originalContent := expr.Name.String()
		if injectInfo.SelectFields == nil {
			injectInfo.SelectFields = make(map[string]common.Expression)
		}
		injectInfo.SelectFields[strings.ToLower(field.Content)] = common.Expression{
			Content:           field.Content,
			OriginalContent:   originalContent,
			Type:              common.EXPR_TYPE_FIELD,
			Alias:             field.Alias,
			IsInAggregateFunc: false,
		}
		return &common.ResolverContent{
			Content:         field.Content,
			OriginalContent: originalContent,
			AliasField:      field.Alias,
		}, nil
	} else {
		if !expr.Qualifier.IsEmpty() {
			return nil, errors.NewParseError("field `%s` not found in dataset '%s'", expr.Name.String(), expr.Qualifier.Name.String())
		} else {

			return nil, errors.NewParseError("field `%s` not found", expr.Name.String())
		}

	}
}
