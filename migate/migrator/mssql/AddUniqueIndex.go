package mssql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

func (m *migratorMssql) GetSqlAddUniqueIndex(db *db.DB, typ reflect.Type) (string, error) {
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
	uk := entityItem.Entity.UniqueConstraints
	for constraintName, cols := range uk {
		var colNames []string
		colNameInConstraint := []string{}

		for _, col := range cols.Cols {
			colNames = append(colNames, m.Quote(col.Name))
			colNameInConstraint = append(colNameInConstraint, col.Name)
		}
		//constraintName := fmt.Sprintf("UQ_%s__%s", entityItem.Entity.TableName, strings.Join(colNameInConstraint, "___"))
		if _, ok := schema.UniqueKeys[strings.ToLower(constraintName)]; !ok {
			constraint := fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)", m.Quote(constraintName), strings.Join(colNames, ", "))
			if strWhere, cols, ok := m.getWhere(cols.Cols); ok {
				sqlCreateConstraint := fmt.Sprintf("CREATE UNIQUE INDEX [%s] ON [%s] (%s) WHERE %s", constraintName, entityItem.Entity.TableName, cols, strWhere)
				sql := fmt.Sprintf(`IF NOT EXISTS (
									SELECT 1
									FROM sys.indexes
									WHERE name = '%s'
									AND object_id = OBJECT_ID('%s')
								)
								BEGIN
									%s;
								END;`, constraintName, entityItem.Entity.TableName, sqlCreateConstraint)
				//sql := fmt.Sprintf("ALTER TABLE [%s] DROP CONSTRAINT [%s];", entityItem.Entity.TableName, constraintName)
				//sql += fmt.Sprintf("CREATE UNIQUE INDEX [%s] ON [%s] (%s) WHERE %s;", constraintName, entityItem.Entity.TableName, cols, strWhere)
				scripts = append(scripts, sql)
			} else {

				script := fmt.Sprintf("ALTER TABLE %s ADD %s", m.Quote(entityItem.Entity.TableName), constraint)
				scripts = append(scripts, script)
			}

		}
	}
	return strings.Join(scripts, ";\n"), nil

}

/*
-- Xóa constraint cũ nếu đã tồn tại
ALTER TABLE [users] DROP CONSTRAINT [UQ_users__uq_email];

-- Tạo unique index có filter
CREATE UNIQUE INDEX [UQ_users__uq_email]
ON [users] ([email])
WHERE [email] IS NOT NULL;
*/
func (m *migratorMssql) getWhere(Cols []entity.ColumnDef) (string, string, bool) {
	items := []string{}
	ok := false
	cols := []string{}
	for _, c := range Cols {
		if c.Nullable {
			ok = true
			items = append(items, m.Quote(c.Name)+" IS NOT NULL")
		}
		cols = append(cols, m.Quote(c.Name))
	}
	return strings.Join(items, " AND "), strings.Join(cols, ","), ok
}
