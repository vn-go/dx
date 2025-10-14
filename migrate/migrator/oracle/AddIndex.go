package oracle

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

func (m *MigratorOracle) GetSqlAddIndex(db *db.DB, typ reflect.Type) (string, error) {
	scripts := []string{}

	// Load database schema hiện tại
	schema, err := m.loader.LoadFullSchema(db)
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

	for _, cols := range entityItem.Entity.IndexConstraints {
		var colNames []string
		var colNameInConstraint []string

		for _, col := range cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}

		constraintName := fmt.Sprintf("IDX_%s__%s", entityItem.Entity.TableName, strings.Join(colNameInConstraint, "_"))

		// Nếu index chưa tồn tại trong schema
		if _, ok := schema.Indexes[constraintName]; !ok {
			// PostgreSQL mặc định dùng BTREE, có thể thêm USING nếu cần
			sql := fmt.Sprintf(
				"CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
				m.Quote(constraintName),
				m.Quote(entityItem.Entity.TableName),
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
