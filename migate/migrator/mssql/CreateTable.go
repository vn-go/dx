package mssql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

/*
Hàm này là 1 implementation của interface .

	type IMigrator interface {
		GetSqlCreateTable(entityType reflect.Type) (string, error)
	}
*/
func (m *migratorMssql) GetSqlCreateTable(db *db.DB, typ reflect.Type) (string, error) {
	mapType := m.GetColumnDataTypeMapping()                      // load mapping data type from migrator
	defaultValueByFromDbTag := m.GetGetDefaultValueByFromDbTag() // load mapping default value from db tag
	schemaLoader := m.GetLoader()                                //<-- get the schema loader injected from the migrator
	if schemaLoader == nil {                                     //<-- make sure schema loader is not nil
		return "", fmt.Errorf("schema loader is nil, please set it by call SetLoader() function in %s", reflect.TypeOf(m).Elem())
	}
	// Load database schema hiện tại
	schema, err := schemaLoader.LoadFullSchema(db) //<-- Load schema from the database. LoadFullSchema is called only once per database
	if err != nil {
		return "", err
	}

	// ModelRegistry is an entity that obviously exists
	// in the eorm library. It allows retrieving information
	// about a model that has been registered by the developer.
	entityItem, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return "", err
	}
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	tableName := entityItem.Entity.TableName

	if _, ok := schema.Tables[strings.ToLower(tableName)]; ok {
		/*
			If the table already exists in the database, there is no need to create it .
		*/
		return "", nil
	}

	strCols := []string{}            //<-- lits of column name in database table
	newTableMap := map[string]bool{} //<-- this variable is used for update schema.Tables map after create table
	for _, col := range entityItem.Entity.Cols {
		newTableMap[strings.ToLower(col.Name)] = true // add column name to newTableMap
		fieldType := col.Field.Type                   // get field type, the type of the field in the struct
		if fieldType.Kind() == reflect.Ptr {          // if field type is pointer, get the real type of the field
			fieldType = fieldType.Elem()
		}

		if _, ok := mapType[fieldType]; !ok {
			/* If there is any error or omission in the mapping declaration between Go data types and dbtypes,
			panic immediately and do not allow any other commands to run. */
			errMsg := fmt.Sprintf("not support field type %s, review GetColumnDataTypeMapping() function in %s ", fieldType.String(), reflect.TypeOf(m).Elem())
			panic(errMsg)
		}
		sqlType := mapType[fieldType]
		if col.Length != nil {
			sqlType = fmt.Sprintf("%s(%d)", sqlType, *col.Length)
		}

		colDef := m.Quote(col.Name) + " " + sqlType

		if col.IsAuto {
			colDef += " IDENTITY(1,1)"
		}

		if col.Nullable {
			colDef += " NULL"
		} else {
			colDef += " NOT NULL"
		}

		if col.Default != "" {
			defaultVal, err := internal.Helper.GetDefaultValue(col.Default, defaultValueByFromDbTag)
			if err != nil {
				err = fmt.Errorf("not support default value from %s, review GetGetDefaultValueByFromDbTag() function in %s ", col.Default, "github.com/vn-go/xdb/migrate/migrator.mssql.CreateTable.go")
				panic(err)
			}

			colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
		}

		strCols = append(strCols, colDef)
	}

	/*
	   Every table created in the database must have a primary key.
	   This part will iterate through the list of keys. Here, the system needs to handle the common case where
	   the table has one or multiple columns as the key.
	*/
	for _, cols := range entityItem.Entity.PrimaryConstraints {
		//var colNames []string
		// colNameInConstraint := []string{}

		var pkCols []string
		var pkColName []string
		for _, col := range cols {
			if col.PKName != "" {
				pkCols = append(pkCols, m.Quote(col.Name))
				pkColName = append(pkColName, col.Name)
			}
		}
		pkConstraintName := ""
		if len(pkCols) > 0 {
			// Constraint name theo chuẩn PK_<table>__<col1>_<col2>
			pkConstraintName = fmt.Sprintf("PK_%s__%s", tableName, strings.Join(pkColName, "_"))
			constraint := fmt.Sprintf("CONSTRAINT %s PRIMARY KEY (%s)", m.Quote(pkConstraintName), strings.Join(pkCols, ", "))
			strCols = append(strCols, constraint)
		}
		// constraintName = fmt.Sprintf("PK_%s__%s", tableName, strings.Join(colNameInConstraint, "___"))
		//constraint := fmt.Sprintf("CONSTRAINT %s PRIMARY KEY (%s)", m.Quote(pkConstraintName), strings.Join(colNames, ", "))
		strCols = append(strCols)

	}

	// Finally, Create Table here
	sql := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", m.Quote(tableName), strings.Join(strCols, ",\n  "))
	schema.Tables[strings.ToLower(tableName)] = newTableMap

	return sql, nil
}
