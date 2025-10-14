package mssql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

func ixIfNoExist(constraintName, tableName, sqlAddIX string) string {
	ret := fmt.Sprintf(`
				IF NOT EXISTS (
				SELECT 1 
				FROM sys.indexes 
				WHERE name = N'%s' 
				AND object_id = OBJECT_ID(N'%s')
			)
			BEGIN
				%s;
			END;
				`, constraintName, tableName, sqlAddIX)
	return ret
}
func (m *migratorMssql) GetSqlAddIndex(db *db.DB, typ reflect.Type) (string, error) {
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
		colNameInConstraint := []string{}
		for _, col := range cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}
		constraintName := fmt.Sprintf("IDX_%s__%s", entityItem.Entity.TableName, strings.Join(colNameInConstraint, "_"))
		if _, ok := schema.Indexes[strings.ToLower(constraintName)]; !ok {
			constraint := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", m.Quote(constraintName), m.Quote(entityItem.Entity.TableName), strings.Join(colNames, ", "))
			constraintIfNotExist := ixIfNoExist(constraintName, entityItem.Entity.TableName, constraint)
			scripts = append(scripts, constraintIfNotExist)

		}
	}
	return strings.Join(scripts, ";\n"), nil

}
