package quicky

import (
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

// sqlNode.BuilSelect
func (s sqlNode) BuilSelect(dialect types.Dialect, textParams []string, args []any, argsInspects *ArgInspects, field *FieldInspects, dict *Dictionanry) (string, error) {
	retItems := []string{}
	for _, x := range s.nodes {

		if fx, ok := x.(*QueryItem); ok {
			sqlN := newSqlParser()
			err := sqlN.Parse(fx, dialect, textParams, args...)
			if err != nil {
				return "", err
			}
			retItems = append(retItems, sqlN.Statement)

			//do something
		} else {
			content, err := resolver.Resolve(x.(sqlparser.SQLNode), dialect, textParams, args, argsInspects, field, dict, C_TYPE_SELECT)
			if err != nil {
				return "", err
			}
			retItems = append(retItems, content)
		}
	}
	return strings.Join(retItems, ","), nil
}
