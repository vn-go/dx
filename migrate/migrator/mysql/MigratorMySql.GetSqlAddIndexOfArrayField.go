package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/entity"
)

// MigratorMySql.GetSqlAddIndexOfArrayField.go
func (m *MigratorMySql) GetSqlAddIndexOfArrayField(db *db.DB, typ reflect.Type, schema string, constraintName string, tableName string, cols []entity.ColumnDef) string {
	/*
		ALTER TABLE departments
		ADD INDEX idx_children_id ((CAST(JSON_EXTRACT(children_id, '$[*]') AS UNSIGNED ARRAY)));
	*/
	items := []string{}
	for _, x := range cols {
		typ := x.Field.Type
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if typ.Kind() == reflect.Slice {
			items = append(items, fmt.Sprintf("(CAST(JSON_EXTRACT(`%s`, '$[*]') AS UNSIGNED ARRAY))", x.Name))
		} else {
			items = append(items, "`"+x.Name+"`")
		}
	}
	sqlRet := fmt.Sprintf("ALTER TABLE `%s` ADD INDEX `%s`(%s)", tableName, constraintName, strings.Join(items, ","))
	return sqlRet
}
