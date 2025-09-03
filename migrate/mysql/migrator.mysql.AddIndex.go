package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/common"
)

func (m *MigratorMySql) GetSqlAddIndex(typ reflect.Type) (string, error) {
	scripts := []string{}

	// Load schema hiện tại
	schema, err := m.Loader.LoadFullSchema(m.Db)
	if err != nil {
		return "", err
	}

	// Lấy entity đã đăng ký
	entityItem := common.ModelRegistry.GetModelByType(typ)
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	for _, cols := range entityItem.Entity.GetIndexConstraints() {
		var colNames []string
		var colNameInConstraint []string

		for _, col := range cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}

		constraintName := fmt.Sprintf("IDX_%s__%s", entityItem.TableName, strings.Join(colNameInConstraint, "_"))

		// Nếu chưa tồn tại index này trong schema
		if _, ok := schema.Indexes[constraintName]; !ok {
			stmt := fmt.Sprintf(
				"CREATE INDEX %s ON %s (%s)",
				m.Quote(constraintName),
				m.Quote(entityItem.TableName),
				strings.Join(colNames, ", "),
			)
			scripts = append(scripts, stmt)
		}
	}

	return strings.Join(scripts, ";\n"), nil
}
