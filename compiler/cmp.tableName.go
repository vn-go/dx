package compiler

import (
	"strings"

	"github.com/vn-go/dx/model"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) tableName(expr sqlparser.TableName, cmpType COMPILER) (string, error) {
	tableNameMatch := strings.ToLower(expr.Name.String())

	ent := model.ModelRegister.FindEntityByName(tableNameMatch)
	if ent != nil {
		if retTable, ok := cmp.dict.TableAlias[tableNameMatch]; ok {
			return cmp.dialect.Quote(ent.TableName) + " " + cmp.dialect.Quote(retTable), nil
		}
		return cmp.dialect.Quote(ent.TableName), nil
	}
	return cmp.dialect.Quote(expr.Name.String()), nil
}
