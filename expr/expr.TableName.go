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
	if context.Alias == nil {
		context.Alias = map[string]string{}
	}

	if context.Purpose == BUILD_JOIN {
		if aliasTableName, ok := context.stackAliasTables.Pop(); ok {
			if _, ok := context.Alias[aliasTableName]; !ok {
				context.Tables = append(context.Tables, aliasTableName)
				context.Alias[aliasTableName] = aliasTableName
				context.aliasToDbTable[aliasTableName] = tableName
			}
			compileTableName := tableName
			if _, ok := (*context.schema)[tableName]; !ok {
				compileTableName = internal.Utils.Pluralize(tableName)
				context.aliasToDbTable[aliasTableName] = tableName

			} else {
				if context.aliasToDbTable == nil {
					context.aliasToDbTable = map[string]string{}
				}
				context.aliasToDbTable[aliasTableName] = tableName
			}
			return context.Dialect.Quote(compileTableName) + " AS " + context.Dialect.Quote(aliasTableName), nil
		} else {

			compileTableName := tableName
			if _, ok := (*context.schema)[tableName]; !ok {
				compileTableName = internal.Utils.Pluralize(tableName)

			}
			if _, ok := context.Alias[tableName]; !ok {
				context.Tables = append(context.Tables, tableName)
				context.Alias[tableName] = "T" + strconv.Itoa(len(context.Tables))
			}
			return context.Dialect.Quote(compileTableName) + " AS " + context.Dialect.Quote(context.Alias[tableName]), nil
		}
	} else {
		if _, ok := (*context.schema)[tableName]; ok {
			return context.Dialect.Quote(tableName), nil
		}
		tableName = internal.Utils.Pluralize(tableName)
		return context.Dialect.Quote(tableName), nil
	}

}
