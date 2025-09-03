package expr

import (
	"strconv"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// ComparisonExpr
func (e *exprReceiver) ColName(context *exprCompileContext, expr sqlparser.ColName) (string, error) {
	if context.aliasToDbTable == nil {
		context.aliasToDbTable = map[string]string{}
	}

	tableName := expr.Qualifier.Name.String()
	fieldName := expr.Name.String()
	aliasFieldName := ""
	if context.purpose == BUILD_SELECT {
		aliasFieldName = expr.Name.String()
	}
	if context.schema == nil {
		context.schema = &map[string]bool{}
	}
	checlAlaisTableName := internal.Utils.Plural(tableName)
	if _, ok := (context.alias)[checlAlaisTableName]; ok { // if not found in database schema, then assume it is a plural table name
		tableName = checlAlaisTableName
	}

	if _, ok := (*context.schema)[tableName]; !ok {
		/*
			if not found in database calculate alias table name , field name and alias field name
		*/

		if _, ok := context.aliasToDbTable[tableName]; ok { //<-- if compiled before for join purpose has alias table name
			fieldName = internal.Utils.ToSnakeCase(fieldName) //<-- 100% sure that field name is in snake case
		}
		if aliasTable, ok := context.alias[tableName]; ok {
			tableName = aliasTable
		} else {
			if context.purpose == BUILD_JOIN {
				/*
					if purpose is join, the compiling process need
					extract tables if they were found when compiling the query
				*/
				context.tables = append(context.tables, tableName)
				context.alias[tableName] = "T" + strconv.Itoa(len(context.tables))
				tableName = context.alias[tableName]
			} else {
				if tableName == "" && len(context.tables) == 1 {
					tableName = context.alias[context.tables[0]]
				} else {
					tableName = internal.Utils.Plural(tableName)
				}
			}
		}

	} else if context.purpose == BUILD_OFFSET {
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
	if context.purpose == BUILD_SELECT {
		/*
			if purpose is select, then return tablename.fieldname as aliasfieldname
			Heed: quote all the things
		*/
		if alias, ok := context.alias[tableName]; ok {
			tableName = alias
			aliasFieldName = fieldName
			fieldName = internal.Utils.ToSnakeCase(fieldName)

		}
		return context.dialect.Quote(tableName, fieldName) + " AS " + context.dialect.Quote(aliasFieldName), nil

	} else {
		if alias, ok := context.alias[tableName]; ok {
			tableName = alias

		}
		fieldName = internal.Utils.ToSnakeCase(fieldName)
		return context.dialect.Quote(tableName, fieldName), nil
	}

	// if not found in database schema, then assume it is a plural table name
	// if context.schema == nil {
	// 	context.schema = &map[string]bool{}
	// }
	// tableName := expr.Qualifier.Name.String()
	// fieldName := expr.Name.String()
	// var fullName string
	// if aliasField, ok := context.stackAliasFields.Pop(); ok {
	// 	if context.purpose == BUILD_SELECT {
	// 		fullName = context.dialect.Quote(tableName, fieldName) + " AS " + context.dialect.Quote(aliasField)
	// 	}
	// } else {
	// 	if context.purpose == BUILD_SELECT {
	// 		fullName = context.dialect.Quote(context.alias[tableName], utils.ToSnakeCase(fieldName)) + " AS " + context.dialect.Quote(expr.Name.String())
	// 	} else {
	// 		fullName = context.dialect.Quote(context.alias[tableName], utils.ToSnakeCase(fieldName))
	// 	}

	// }
	// return fullName, nil
	// if context.purpose == BUILD_SELECT {
	// 	return fullName + " AS " + context.dialect.Quote(expr.Name.String()), nil

	// } else if context.purpose == BUILD_SELECT {
	// 	return fullName, nil
	// } else {
	// 	return fullName, nil
	// }

	// if _, ok := context.alias[tableName]; !ok { // if not found in alias, then check if it is a schema table
	// 	if _, ok := (*context.schema)[tableName]; !ok { // if not found in database schema, then assume it is a plural table name

	// 		tableName = utils.Plural(tableName)
	// 		//fieldName = utils.ToSnakeCase(expr.Name.String())
	// 	}
	// 	if _, ok := context.alias[tableName]; !ok {

	// 		context.tables = append(context.tables, tableName)
	// 		context.alias[tableName] = "T" + strconv.Itoa(len(context.tables))
	// 	}
	// }
	// if context.purpose == BUILD_FUNC || context.purpose == BUILD_JOIN {
	// 	compileTableName := context.pluralTableName(tableName)
	// 	compileFieldName := utils.ToSnakeCase(expr.Name.String())

	// 	return context.dialect.Quote(compileTableName, compileFieldName), nil
	// }
	// if context.purpose == BUILD_SELECT {
	// 	if aliasField, ok := context.stackAliasFields.Pop(); ok {

	// 		compileFieldName := utils.ToSnakeCase(expr.Name.String())
	// 		ret := context.dialect.Quote(tableName, compileFieldName) + " AS " + context.dialect.Quote(aliasField)

	// 		return ret, nil
	// 	}
	// 	compileTableName := tableName
	// 	if aliasTable, ok := context.alias[tableName]; ok {
	// 		compileTableName = aliasTable
	// 	}
	// 	compileTableName = context.pluralTableName(tableName)
	// 	compileFieldName := utils.ToSnakeCase(expr.Name.String())
	// 	return context.dialect.Quote(compileTableName, compileFieldName) + " AS " + context.dialect.Quote(expr.Name.String()), nil
	// }
	// compileTableName := context.pluralTableName(tableName)
	// compileFieldName := utils.ToSnakeCase(expr.Name.String())
	// return context.dialect.Quote(compileTableName, compileFieldName), nil

}
