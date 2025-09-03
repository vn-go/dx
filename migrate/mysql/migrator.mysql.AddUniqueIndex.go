package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/common"
)

func (m *MigratorMySql) GetSqlAddUniqueIndex(typ reflect.Type) (string, error) {
	scripts := []string{}

	// Load database schema hiện tại
	schema, err := m.Loader.LoadFullSchema(m.Db)
	if err != nil {
		return "", err
	}

	// Lấy entity đã đăng ký
	entityItem := common.ModelRegistry.GetModelByType(typ)
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	uk := entityItem.Entity.GetUniqueConstraints()

	for _, cols := range uk {
		var colNames []string
		var colNameInConstraint []string
		for _, col := range cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}

		constraintName := fmt.Sprintf("UQ_%s__%s", entityItem.TableName, strings.Join(colNameInConstraint, "___"))

		if _, ok := schema.UniqueKeys[constraintName]; !ok {
			script := fmt.Sprintf(
				"CREATE UNIQUE INDEX %s ON %s (%s)",
				m.Quote(constraintName),
				m.Quote(entityItem.TableName),
				strings.Join(colNames, ", "),
			)
			scripts = append(scripts, script)
		}
	}

	return strings.Join(scripts, ";\n"), nil
}
