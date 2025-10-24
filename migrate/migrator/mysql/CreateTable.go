package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/migrate/loader/types"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (m *MigratorMySql) GetSqlCreateTable(db *db.DB, typ reflect.Type, schema string) (string, error) {
	mapType := m.GetColumnDataTypeMapping()
	defaultValueByFromDbTag := m.GetGetDefaultValueByFromDbTag()
	schemaLoader := m.GetLoader()
	if schemaLoader == nil {
		return "", fmt.Errorf("schema loader is nil, please set it by call SetLoader() function in %s", reflect.TypeOf(m).Elem())
	}

	schemaData, err := schemaLoader.LoadFullSchema(db, schema)
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
	if _, ok := schemaData.Tables[tableName]; ok {
		return "", nil // table already exists
	}

	strCols := []string{}
	newTableMap := map[string]bool{}
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

		if col.Length != nil && (col.Field.Type == reflect.TypeFor[string]() || col.Field.Type == reflect.TypeFor[*string]()) {
			sqlType = fmt.Sprintf("%s(%d)", sqlType, *col.Length)
		}
		if strings.ToLower(sqlType) == "varchar" {
			sqlType = "text"
		}

		colDef := m.Quote(col.Name) + " " + sqlType

		if col.IsAuto {
			colDef += " AUTO_INCREMENT"
		}

		if col.Nullable {
			colDef += " NULL"
		} else {
			colDef += " NOT NULL"
		}

		if col.Default != "" {
			defaultVal, err := internal.Helper.GetDefaultValue(col.Default, defaultValueByFromDbTag)

			if err != nil {
				err = fmt.Errorf("not support default value from %s, review GetGetDefaultValueByFromDbTag() function in %s ", col.Default, "github.com/vn-go/xdb/migrate/migrator.mysql.CreateTable.go")
				panic(err)
			}
			if defaultVal != internal.Helper.SkipDefaulValue {
				colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
			}

		}

		strCols = append(strCols, colDef)
	}

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
			// MySQL thường không cần đặt tên, nhưng bạn vẫn có thể nếu muốn
			constraint := fmt.Sprintf("CONSTRAINT %s PRIMARY KEY (%s)", m.Quote(pkConstraintName), strings.Join(pkCols, ", "))
			strCols = append(strCols, constraint)
		}
	}

	sql := fmt.Sprintf("CREATE TABLE %s (\n  %s\n)", m.Quote(tableName), strings.Join(strCols, ",\n  "))
	if !types.SkipLoadSchemaOnMigrate {
		schemaData.Tables[tableName] = newTableMap

	}

	return sql, nil
}
