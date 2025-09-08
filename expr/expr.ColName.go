package expr

import (
	"strconv"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// ComparisonExpr
func (e *exprReceiver) ColName(context *exprCompileContext, expr sqlparser.ColName) (string, error) {
	if context.AliasToDbTable == nil {
		context.AliasToDbTable = map[string]string{}
	}

	tableName := expr.Qualifier.Name.String()
	if context.AlterTableJoin != nil {
		if t, ok := context.AlterTableJoin[tableName]; ok {
			tableName = t
		}
	}
	fieldName := expr.Name.String()
	aliasFieldName := ""
	if context.Purpose == BUILD_SELECT {
		var found bool
		if aliasFieldName, found = context.stackAliasFields.Pop(); !found {
			aliasFieldName = expr.Name.String()
		}

	}
	if context.schema == nil {
		context.schema = &map[string]bool{}
	}
	checklAlaisTableName := internal.Utils.Pluralize(tableName)
	if _, ok := (context.Alias)[checklAlaisTableName]; ok { // if not found in database schema, then assume it is a plural table name
		tableName = checklAlaisTableName
	}

	if _, ok := (*context.schema)[tableName]; !ok {
		/*
			if not found in database calculate alias table name , field name and alias field name
		*/

		if _, ok := context.AliasToDbTable[tableName]; ok { //<-- if compiled before for join purpose has alias table name
			fieldName = internal.Utils.SnakeCase(fieldName) //<-- 100% sure that field name is in snake case
		}
		if aliasTable, ok := context.Alias[tableName]; ok {
			tableName = aliasTable
		} else {
			if context.Purpose == BUILD_JOIN {
				/*
					if purpose is join, the compiling process need
					extract tables if they were found when compiling the query
				*/
				context.Tables = append(context.Tables, tableName)
				context.Alias[tableName] = "T" + strconv.Itoa(len(context.Tables))
				tableName = context.Alias[tableName]
			} else {
				if tableName == "" && len(context.Tables) == 1 {
					tableName = context.Alias[context.Tables[0]]
				} else {
					tableName = internal.Utils.Pluralize(tableName)
				}
			}
		}

	} else if context.Purpose == BUILD_OFFSET {
		/*
			if found in database calculate alias field name
			tableName from database schema
			column name no change because it is already in SQL statement where DEV declared in their code
		*/
		/*
			But important that we need to check if alias field name is already in stack, if yes then use it, otherwise use field name
		*/
		if aliasFieldFromStack, ok := context.stackAliasFields.Pop(); ok {
			aliasFieldName = aliasFieldFromStack
		} else {
			aliasFieldName = fieldName
		}

	}
	if context.Purpose == BUILD_SELECT {
		/*
			if purpose is select, then return tablename.fieldname as aliasfieldname
			Heed: quote all the things
		*/
		if alias, ok := context.Alias[tableName]; ok {
			tableName = alias

			fieldName = internal.Utils.SnakeCase(fieldName)

		}
		return context.Dialect.Quote(tableName, fieldName) + " AS " + context.Dialect.Quote(aliasFieldName), nil

	} else {
		if context.Purpose == BUILD_JOIN {
			if _, ok := context.Alias[tableName]; !ok {
				context.Tables = append(context.Tables, tableName)
				context.Alias[tableName] = tableName
				context.AliasToDbTable[tableName] = tableName
			}

		}
		if alias, ok := context.Alias[tableName]; ok {
			tableName = alias

		}
		fieldName = internal.Utils.SnakeCase(fieldName)
		return context.Dialect.Quote(tableName, fieldName), nil
	}

	// if not found in database schema, then assume it is a plural table name
	// if context.schema == nil {
	// 	context.schema = &map[string]bool{}
	// }
	// tableName := expr.Qualifier.Name.String()
	// fieldName := expr.Name.String()
	// var fullName string
	// if aliasField, ok := context.stackAliasFields.Pop(); ok {
	// 	if context.Purpose == BUILD_SELECT {
	// 		fullName = context.Dialect.Quote(tableName, fieldName) + " AS " + context.Dialect.Quote(aliasField)
	// 	}
	// } else {
	// 	if context.Purpose == BUILD_SELECT {
	// 		fullName = context.Dialect.Quote(context.alias[tableName], internal.Utils.SnakeCase(fieldName)) + " AS " + context.Dialect.Quote(expr.Name.String())
	// 	} else {
	// 		fullName = context.Dialect.Quote(context.alias[tableName], internal.Utils.SnakeCase(fieldName))
	// 	}

	// }
	// return fullName, nil
	// if context.Purpose == BUILD_SELECT {
	// 	return fullName + " AS " + context.Dialect.Quote(expr.Name.String()), nil

	// } else if context.Purpose == BUILD_SELECT {
	// 	return fullName, nil
	// } else {
	// 	return fullName, nil
	// }

	// if _, ok := context.alias[tableName]; !ok { // if not found in alias, then check if it is a schema table
	// 	if _, ok := (*context.schema)[tableName]; !ok { // if not found in database schema, then assume it is a plural table name

	// 		tableName = utils.Plural(tableName)
	// 		//fieldName = internal.Utils.SnakeCase(expr.Name.String())
	// 	}
	// 	if _, ok := context.alias[tableName]; !ok {

	// 		context.tables = append(context.tables, tableName)
	// 		context.alias[tableName] = "T" + strconv.Itoa(len(context.tables))
	// 	}
	// }
	// if context.Purpose == BUILD_FUNC || context.Purpose == BUILD_JOIN {
	// 	compileTableName := context.pluralTableName(tableName)
	// 	compileFieldName := internal.Utils.SnakeCase(expr.Name.String())

	// 	return context.Dialect.Quote(compileTableName, compileFieldName), nil
	// }
	// if context.Purpose == BUILD_SELECT {
	// 	if aliasField, ok := context.stackAliasFields.Pop(); ok {

	// 		compileFieldName := internal.Utils.SnakeCase(expr.Name.String())
	// 		ret := context.Dialect.Quote(tableName, compileFieldName) + " AS " + context.Dialect.Quote(aliasField)

	// 		return ret, nil
	// 	}
	// 	compileTableName := tableName
	// 	if aliasTable, ok := context.alias[tableName]; ok {
	// 		compileTableName = aliasTable
	// 	}
	// 	compileTableName = context.pluralTableName(tableName)
	// 	compileFieldName := internal.Utils.SnakeCase(expr.Name.String())
	// 	return context.Dialect.Quote(compileTableName, compileFieldName) + " AS " + context.Dialect.Quote(expr.Name.String()), nil
	// }
	// compileTableName := context.pluralTableName(tableName)
	// compileFieldName := internal.Utils.SnakeCase(expr.Name.String())
	// return context.Dialect.Quote(compileTableName, compileFieldName), nil

}
