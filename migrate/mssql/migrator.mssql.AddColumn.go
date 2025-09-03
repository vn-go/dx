package mssql

import (
	"fmt"
	"reflect"
	"strings"

	common "github.com/vn-go/dx/migrate/common"
)

func (m *MigratorMssql) GetSqlAddColumn(typ reflect.Type) (string, error) {
	mapType := m.GetColumnDataTypeMapping()
	defaultValueByFromDbTag := m.GetGetDefaultValueByFromDbTag()

	// Load database schema hiện tại
	schema, err := m.Loader.LoadFullSchema(m.Db)
	if err != nil {
		return "", err
	}

	// Lấy entity đã đăng ký
	entityItem := common.ModelRegistry.GetModelByType(typ)
	if entityItem == nil {
		return "", NewModelError(typ)
	}
	scripts := []string{}
	for _, col := range entityItem.Entity.Cols {
		if _, ok := schema.Tables[entityItem.TableName][col.Name]; !ok {
			fieldType := col.Field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
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
				df, err := common.TypeUtils.GetDefaultValue(col.Default, defaultValueByFromDbTag)
				if err != nil {
					err = fmt.Errorf("not support default value from %s, review GetGetDefaultValueByFromDbTag() function in %s ", col.Default, "vdb/migrate/migrator.mssql.AddColumn.go")
					panic(err)
				}
				colDef += " DEFAULT " + df

				colDef += fmt.Sprintf(" DEFAULT %s", colDef)
			}

			scripts = append(scripts, fmt.Sprintf("ALTER TABLE %s ADD %s", m.Quote(entityItem.TableName), colDef))

			schema.Tables[entityItem.TableName][col.Name] = true
		}
	}

	return strings.Join(scripts, ";\n"), nil

}
