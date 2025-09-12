package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/migate/loader/types"
	"github.com/vn-go/dx/model"
)

/*
The function performs the role of converting or processing data according to the defined interface.
Specifically, it implements the methods declared by the interface, enabling consistent handling or transformation of data based on the interface's contract.
This ensures that any type implementing the interface can be used interchangeably, allowing polymorphic behavior and standardized data processing within the system.

	type IMigrator interface {
		GetSqlCreateTable(entityType reflect.Type) (string, error)
	}
*/
func (m *MigratorPostgres) GetSqlCreateTable(db *db.DB, typ reflect.Type) (string, error) {
	mapType := m.GetColumnDataTypeMapping()
	defaultValueByFromDbTag := m.GetGetDefaultValueByFromDbTag()
	schemaLoader := m.GetLoader()
	if schemaLoader == nil {
		return "", fmt.Errorf("schema loader is nil, please set it by call SetLoader() function in %s", reflect.TypeOf(m).Elem())
	}

	schema, err := schemaLoader.LoadFullSchema(db)
	if err != nil {
		return "", err
	}

	entityItem, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return "", err
	}
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	tableName := entityItem.Entity.TableName
	if _, ok := schema.Tables[tableName]; ok {
		return "", nil
	}

	strCols := []string{}
	newTableMap := map[string]bool{}
	//SequenceScript := map[string]string{}
	scriptSqlCheckLength := []string{}

	for _, col := range entityItem.Entity.Cols {
		newTableMap[col.Name] = true
		fieldType := col.Field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		sqlType, ok := mapType[fieldType]
		if !ok {
			panic(fmt.Sprintf("not support field type %s, review GetColumnDataTypeMapping() function in %s", fieldType.String(), reflect.TypeOf(m).Elem()))
		}

		if col.Length != nil {

			/*ADD CONSTRAINT chk_email_length CHECK (char_length(email) <= 255);*/
			scriptSqlCheckLength = append(scriptSqlCheckLength, m.createCheckLenConstraint(tableName, col))

			// sqlType = fmt.Sprintf("%s(%d)", sqlType, *col.Length)
		}

		colDef := m.Quote(col.Name)

		// Xử lý trường tự động tăng (PostgreSQL)
		if col.IsAuto {

			if fieldType.Kind() == reflect.Int || fieldType.Kind() == reflect.Int64 {
				colDef += " BIGSERIAL"
			} else {
				version, err := db.GetMajorVersion()
				if err != nil {
					colDef += " SERIAL"
				} else {
					if version > 10 {
						colDef += sqlType + " GENERATED ALWAYS AS IDENTITY"
					} else {
						colDef += " SERIAL"
					}
				}

			}
		} else {
			colDef += " " + sqlType
		}

		if !col.Nullable {
			colDef += " NOT NULL"
		}

		if col.Default != "" {
			defaultVal, err := internal.Helper.GetDefaultValue(col.Default, defaultValueByFromDbTag)
			if err != nil {
				err = fmt.Errorf("not support default value from %s, review GetGetDefaultValueByFromDbTag() function in %s ", col.Default, "github.com/vn-go/xdb/migrate/migrator.postgres.GetSqlCreateTable.go")
				panic(err)
			}

			colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
		}

		strCols = append(strCols, colDef)
	}

	// Xử lý khóa chính
	for _, cols := range entityItem.Entity.PrimaryConstraints {

		var pkCols []string
		var pkColNames []string

		for _, col := range cols {
			if col.PKName != "" {
				pkCols = append(pkCols, m.Quote(col.Name))
				pkColNames = append(pkColNames, col.Name)
			}
		}

		if len(pkCols) > 0 {
			pkConstraintName := fmt.Sprintf("PK_%s__%s", tableName, strings.Join(pkColNames, "_"))
			constraint := fmt.Sprintf("CONSTRAINT %s PRIMARY KEY (%s)", m.Quote(pkConstraintName), strings.Join(pkCols, ", "))
			strCols = append(strCols, constraint)
		}
	}

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n);", m.Quote(tableName), strings.Join(strCols, ",\n  "))
	if !types.SkipLoadSchemaOnMigrate {
		schema.Tables[tableName] = newTableMap
	}

	sqlCiText := "CREATE EXTENSION IF NOT EXISTS citext"
	return sqlCiText + ";\n" + sql + strings.Join(scriptSqlCheckLength, ";\n"), nil
}
