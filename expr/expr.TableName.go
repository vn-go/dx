package expr

import (
	"strconv"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (compiler *exprReceiver) TableName(context *exprCompileContext, expr *sqlparser.TableName) (string, error) {

	tableName := expr.Name.String()
	if context.schema == nil {
		context.schema = &map[string]bool{}
	}
	if context.alias == nil {
		context.alias = map[string]string{}
	}

	if context.Purpose == BUILD_JOIN {
		if aliasTableName, ok := context.stackAliasTables.Pop(); ok {
			if _, ok := context.alias[aliasTableName]; !ok {
				context.tables = append(context.tables, aliasTableName)
				context.alias[aliasTableName] = aliasTableName
				context.aliasToDbTable[aliasTableName] = tableName
			}
			compileTableName := tableName
			if _, ok := (*context.schema)[tableName]; !ok {
				compileTableName = internal.Utils.Plural(tableName)
				context.aliasToDbTable[aliasTableName] = tableName

			} else {
				if context.aliasToDbTable == nil {
					context.aliasToDbTable = map[string]string{}
				}
				context.aliasToDbTable[aliasTableName] = tableName
			}
			return context.dialect.Quote(compileTableName) + " AS " + context.dialect.Quote(aliasTableName), nil
		} else {

			compileTableName := tableName
			if _, ok := (*context.schema)[tableName]; !ok {
				compileTableName = internal.Utils.Plural(tableName)

			}
			if _, ok := context.alias[tableName]; !ok {
				context.tables = append(context.tables, tableName)
				context.alias[tableName] = "T" + strconv.Itoa(len(context.tables))
			}
			return context.dialect.Quote(compileTableName) + " AS " + context.dialect.Quote(context.alias[tableName]), nil
		}
	} else {
		if _, ok := (*context.schema)[tableName]; ok {
			return context.dialect.Quote(tableName), nil
		}
		tableName = internal.Utils.Plural(tableName)
		return context.dialect.Quote(tableName), nil
	}

}
