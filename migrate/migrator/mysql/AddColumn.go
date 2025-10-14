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

func (m *MigratorMySql) GetSqlAddColumn(db *db.DB, typ reflect.Type, schema string) (string, error) {
	mapType := m.GetColumnDataTypeMapping()
	defaultValueByFromDbTag := m.GetGetDefaultValueByFromDbTag()

	schemaData, err := m.loader.LoadFullSchema(db, schema)
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

	scripts := []string{}
	tableName := entityItem.Entity.TableName

	for _, col := range entityItem.Entity.Cols {
		if _, ok := schemaData.Tables[tableName][col.Name]; !ok {
			fieldType := col.Field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}

			sqlType, ok := mapType[fieldType]
			if !ok {
				panic(fmt.Sprintf("unsupported field type %s, check GetColumnDataTypeMapping()", fieldType.String()))
			}

			if col.Length != nil {
				sqlType = fmt.Sprintf("%s(%d)", sqlType, *col.Length)
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
				if internal.Helper.IsFloatNumber(col.Default) {
					colDef += fmt.Sprintf(" DEFAULT %s", col.Default)

				} else if internal.Helper.IsNumber(col.Default) {
					colDef += fmt.Sprintf(" DEFAULT %s", col.Default)

				} else if internal.Helper.IsString(col.Default) {
					colDef += fmt.Sprintf(" DEFAULT %s", col.Default)

				} else if val, ok := defaultValueByFromDbTag[col.Default]; ok {
					if val != internal.Helper.SkipDefaulValue {
						colDef += fmt.Sprintf(" DEFAULT %s", val)
					}
				} else {
					panic(fmt.Errorf("unsupported default value from %s, check GetGetDefaultValueByFromDbTag()", col.Default))
				}
			}

			stmt := fmt.Sprintf("ALTER TABLE %s ADD %s", m.Quote(tableName), colDef)
			scripts = append(scripts, stmt)
			if !types.SkipLoadSchemaOnMigrate {
				schemaData.Tables[tableName][col.Name] = true
			}
			// Update schema cache

		}
	}

	return strings.Join(scripts, ";\n"), nil
}
