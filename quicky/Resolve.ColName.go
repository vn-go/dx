package quicky

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

func (r *Resolve) ColName(n *sqlparser.ColName, dialect types.Dialect, textParams []string, dynamicArgs []any, arg *ArgInspects, field *FieldInspects, dict *Dictionanry, cmpType C_TYPE) (string, error) {
	if len(dict.AliasMap) > 1 {
		if n.Qualifier.IsEmpty() {
			return "", newParseError("'%s' requires dataset qualifier", n.Name)
		}
	}
	tableName := n.Qualifier.Name.String()
	fieldName := n.Name.String()
	key := strings.ToLower(fmt.Sprintf("%s.%s", tableName, fieldName))
	content, ok := dict.FieldMap[key]
	if !ok {
		return "", newParseError("'%s' is not a valid field", n.Name)
	}
	(*field)[key] = FieldInspect{
		Expression: content.Content,
		Alias:      content.Alias,
	}
	dict.FieldMap[content.Content] = DictionanryItem{
		Alias: content.Alias,
	}
	return content.Content, nil

}
