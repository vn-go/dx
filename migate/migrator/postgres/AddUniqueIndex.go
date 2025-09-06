package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

func (m *MigratorPostgres) GetSqlAddUniqueIndex(db *db.DB, typ reflect.Type) (string, error) {
	scripts := []string{}

	// Load current schema
	schema, err := m.loader.LoadFullSchema(db)
	if err != nil {
		return "", err
	}

	// Get registered entity
	entityItem, err := model.ModelRegister.GetModelByType(typ)
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	// Duyệt các unique constraint
	for constraintName, cols := range entityItem.Entity.UniqueConstraints {
		var colNames []string
		var colNameInConstraint []string

		for _, col := range cols.Cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}

		//constraintName := fmt.Sprintf("UQ_%s__%s", entityItem.Entity.TableName, strings.Join(colNameInConstraint, "_"))

		// Nếu chưa có trong schema
		if _, ok := schema.UniqueKeys[constraintName]; !ok {
			sql := fmt.Sprintf(
				`ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)`,
				m.Quote(entityItem.Entity.TableName),
				m.Quote(constraintName),
				strings.Join(colNames, ", "),
			)
			scripts = append(scripts, sql)
		}
	}

	if len(scripts) == 0 {
		return "", nil
	}

	return strings.Join(scripts, ";\n") + ";", nil
}
