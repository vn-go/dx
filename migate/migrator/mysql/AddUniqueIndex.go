package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

func (m *MigratorMySql) GetSqlAddUniqueIndex(typ reflect.Type) (string, error) {
	scripts := []string{}

	// Load database schema hiện tại
	schema, err := m.loader.LoadFullSchema()
	if err != nil {
		return "", err
	}

	// Lấy entity đã đăng ký
	entityItem, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return "", err
	}
	if entityItem == nil {
		return "", errors.NewModelError(typ)
	}

	uk := entityItem.Entity.UniqueConstraints

	for constraintName, cols := range uk {
		var colNames []string
		var colNameInConstraint []string
		for _, col := range cols.Cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}

		//constraintName := fmt.Sprintf("UQ_%s__%s", entityItem.Entity.TableName, strings.Join(colNameInConstraint, "___"))

		if _, ok := schema.UniqueKeys[constraintName]; !ok {
			script := fmt.Sprintf(
				"CREATE UNIQUE INDEX %s ON %s (%s)",
				m.Quote(constraintName),
				m.Quote(entityItem.Entity.TableName),
				strings.Join(colNames, ", "),
			)
			scripts = append(scripts, script)
		}
	}

	return strings.Join(scripts, ";\n"), nil
}
