package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/migate/loader/types"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

func (m *MigratorPostgres) GetSqlAddColumn(db *db.DB, typ reflect.Type) (string, error) {
	mapType := m.GetColumnDataTypeMapping()
	defaultValueByFromDbTag := m.GetGetDefaultValueByFromDbTag()

	// Load current schema
	schema, err := m.loader.LoadFullSchema(db)
	if err != nil {
		return "", err
	}

	// Get registered model
	entityItem, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return "", err
	}
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	scripts := []string{}
	checkLengthScripts := []string{}

	for _, col := range entityItem.Entity.Cols {
		// Column chưa tồn tại thì mới thêm
		if _, ok := schema.Tables[entityItem.Entity.TableName][col.Name]; !ok {
			fieldType := col.Field.Type
			if fieldType.Kind() == reflect.Ptr {
				fieldType = fieldType.Elem()
			}

			sqlType, ok := mapType[fieldType]
			if !ok {
				return "", fmt.Errorf("unsupported field type %s, check GetColumnDataTypeMapping", fieldType.String())
			}

			if col.Length != nil {
				strCheck := m.createCheckLenConstraint(entityItem.Entity.TableName, col)
				if strCheck != "" {
					checkLengthScripts = append(checkLengthScripts, strCheck)
				}

			}

			colDef := m.Quote(col.Name)

			// Xử lý auto increment trong PostgreSQL
			if col.IsAuto {
				if fieldType.Kind() == reflect.Int || fieldType.Kind() == reflect.Int64 {
					colDef += " BIGSERIAL"
				} else {
					colDef += fmt.Sprintf(" %s GENERATED ALWAYS AS IDENTITY", sqlType)
				}
			} else {
				colDef += " " + sqlType
			}

			if !col.Nullable {
				colDef += " NOT NULL"
			}

			if col.Default != "" {
				defaultVal := ""
				if val, ok := defaultValueByFromDbTag[col.Default]; ok {
					defaultVal = val
				} else {
					panic(fmt.Errorf("unsupported default tag: %s in %s", col.Default, reflect.TypeOf(m).Elem()))
				}
				colDef += fmt.Sprintf(" DEFAULT %s", defaultVal)
			}

			script := fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s", m.Quote(entityItem.Entity.TableName), colDef)
			scripts = append(scripts, script)

			// Update schema cache
			if !types.SkipLoadSchemaOnMigrate {
				schema.Tables[entityItem.Entity.TableName][col.Name] = true
			}

		}
	}

	if len(scripts) == 0 {
		return "", nil
	}
	scripts = append(scripts, checkLengthScripts...)

	return strings.Join(scripts, ";\n"), nil
}
