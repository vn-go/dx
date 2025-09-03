package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/common"
)

func (m *MigratorPostgres) GetSqlAddUniqueIndex(typ reflect.Type) (string, error) {
	scripts := []string{}

	// Load current schema
	schema, err := m.Loader.LoadFullSchema(m.Db)
	if err != nil {
		return "", err
	}

	// Get registered entity
	entityItem := common.ModelRegistry.GetModelByType(typ)
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	// Duyệt các unique constraint
	for _, cols := range entityItem.Entity.GetUniqueConstraints() {
		var colNames []string
		var colNameInConstraint []string

		for _, col := range cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}

		constraintName := fmt.Sprintf("UQ_%s__%s", entityItem.TableName, strings.Join(colNameInConstraint, "_"))

		// Nếu chưa có trong schema
		if _, ok := schema.UniqueKeys[constraintName]; !ok {
			sql := fmt.Sprintf(
				`ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)`,
				m.Quote(entityItem.TableName),
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
