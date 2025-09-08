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
	} else {
		if expr.Qualifier.IsEmpty() {
			return "", fmt.Errorf("field %s must qualify table by the following tables: %s\nSQL:\n%s", expr.Name.String(), strings.Join(cmp.dict.Tables, ","), cmp.sql)

		} else {
			tableName := expr.Qualifier.Name.String()
			fieldName := expr.Name.String()
			fieldMatch := strings.ToLower(fmt.Sprintf("%s.%s", tableName, fieldName))
			if ret, ok := cmp.dict.Field[fieldMatch]; ok {
				return ret, nil
			} else {
				if tableAlias, ok := cmp.dict.TableAlias[strings.ToLower(tableName)]; ok {
					return cmp.dialect.Quote(tableAlias, fieldName), nil
				}
				return cmp.dialect.Quote(tableName, fieldName), nil
			}

		}
	}

}
