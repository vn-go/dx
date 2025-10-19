package quicky

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

func (r *Resolve) ColName(n *sqlparser.ColName, dialect types.Dialect, textParams []string, dynamicArgs []any, arg *ArgInspects, field *FieldInspects, dict *Dictionanry, cmpType C_TYPE) (string, error) {
	tableName := n.Qualifier.Name.String()
	if n.Qualifier.IsEmpty() {
		if len(dict.Entities) > 1 {
			return "", newParseError("'%s' requires dataset qualifier", n.Name)
		}
		for _, x := range dict.Entities {
			tableName = x.EntityType.Name()
		}

	}
	fieldName := n.Name.String()
	key := strings.ToLower(fmt.Sprintf("%s.%s", tableName, fieldName))
	content, ok := dict.FieldMap[key]
	if !ok {
		return "", newParseError("'%s' was not found in dataset '%s'", n.Name, tableName)
	}
	(*field)[key] = FieldInspect{
		Expression: content.Content,
		Alias:      content.Alias,
	}
	(*field)[content.Content] = (*field)[key]
	dict.FieldMap[content.Content] = DictionanryItem{
		Alias: content.Alias,
	}
	return content.Content, nil

}
