package mssql

import (
	"fmt"
	"reflect"
	"strings"

	common "github.com/vn-go/dx/migrate/common"
)

var NewModelError func(typ reflect.Type) error

func (m *MigratorMssql) GetSqlAddIndex(typ reflect.Type) (string, error) {
	scripts := []string{}

	// Load database schema hiện tại
	schema, err := m.Loader.LoadFullSchema(m.Db)
	if err != nil {
		return "", err
	}
	fmt.Println(typ.String())
	// Lấy entity đã đăng ký
	entityItem := common.ModelRegistry.GetModelByType(typ)
	if entityItem == nil {
		return "", NewModelError(typ)
	}
	for _, cols := range entityItem.Entity.GetIndexConstraints() {
		var colNames []string
		colNameInConstraint := []string{}
		for _, col := range cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}
		constraintName := fmt.Sprintf("IDX_%s__%s", entityItem.TableName, strings.Join(colNameInConstraint, "_"))
		if _, ok := schema.UniqueKeys[constraintName]; !ok {
			constraint := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", m.Quote(constraintName), m.Quote(entityItem.TableName), strings.Join(colNames, ", "))
			scripts = append(scripts, constraint)

		}
	}
	return strings.Join(scripts, ";\n"), nil

}
