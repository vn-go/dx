package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) colName(expr *sqlparser.ColName, cmpType COMPILER) (string, error) {
	if len(cmp.dict.Tables) == 1 {
		field := expr.Name.String()
		matchField := strings.ToLower(fmt.Sprintf("%s.%s", cmp.dict.Tables[0], field))

		if retField, ok := cmp.dict.Field[matchField]; ok {
			return retField, nil
		}
		alias := cmp.dict.TableAlias[cmp.dict.Tables[0]]
		return cmp.dialect.Quote(alias) + "." + cmp.dialect.Quote(expr.Name.String()), nil
	}
	return cmp.dialect.Quote(expr.Name.String()), nil
}
