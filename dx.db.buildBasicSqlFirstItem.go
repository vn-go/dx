package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (db *DB) buildBasicSqlFirstItem(typ reflect.Type, filter string) (string, error) {
	key := db.DriverName + "://" + db.DbName + "/" + typ.String() + "/buildBasicSqlFirstItem/" + filter
	return internal.OnceCall(key, func() (string, error) {
		dialect := factory.DialectFactory.Create(db.DriverName)

		repoType, err := model.ModelRegister.GetModelByType(typ)
		if err != nil {
			return "", err
		}
		tableName := repoType.Entity.TableName
		compiler, err := expr.CompileJoin(tableName, db.DB)
		if err != nil {
			return "", err
		}
		tableName = compiler.Content
		columns := repoType.Entity.Cols

		fieldsSelect := make([]string, len(columns))
		for i, col := range columns {
			fieldsSelect[i] = repoType.Entity.TableName + "." + col.Field.Name + " AS " + col.Field.Name
		}
		compiler.Context.Purpose = expr.BUILD_SELECT
		err = compiler.BuildSelectField(strings.Join(fieldsSelect, ", "))
		if err != nil {
			return "", err
		}
		strField := compiler.Content

		sql := fmt.Sprintf("SELECT %s FROM %s", strField, tableName)
		if filter != "" {
			compiler.Context.Purpose = expr.BUILD_WHERE
			err = compiler.BuildWhere(filter)
			if err != nil {
				return "", err
			}
			sql += " WHERE " + compiler.Content
		}
		var limitPtr *uint64 
		sql = dialect.LimitAndOffset(sql, nil, limitPtr, "")
		return sql, nil
	})

}
func (db *DB) buildBasicSqlFindItem(typ reflect.Type, filter string, orderStr string) (string, error) {
	key := db.DriverName + "://" + db.DbName + "/" + typ.String() + "/buildBasicSqlFindItem/" + filter + "/" + orderStr
	return internal.OnceCall(key, func() (string, error) {

		repoType, err := model.ModelRegister.GetModelByType(typ)
		if err != nil {
			return "", err
		}
		tableName := repoType.Entity.TableName
		compiler, err := expr.CompileJoin(tableName, db.DB)
		if err != nil {
			return "", err
		}
		tableName = compiler.Content
		columns := repoType.Entity.Cols

		fieldsSelect := make([]string, len(columns))
		for i, col := range columns {
			fieldsSelect[i] = repoType.Entity.TableName + "." + col.Field.Name + " AS " + col.Field.Name
		}
		compiler.Context.Purpose = expr.BUILD_SELECT
		err = compiler.BuildSelectField(strings.Join(fieldsSelect, ", "))
		if err != nil {
			return "", err
		}
		strField := compiler.Content

		sql := fmt.Sprintf("SELECT %s FROM %s", strField, tableName)
		if filter != "" {
			compiler.Context.Purpose = expr.BUILD_WHERE
			err = compiler.BuildWhere(filter)
			if err != nil {
				return "", err
			}
			sql += " WHERE " + compiler.Content
		}
		if orderStr != "" {
			compiler.Context.Purpose = expr.BUILD_ORDER
			err = compiler.BuildSortField(orderStr)
			if err != nil {
				return "", err
			}
			sql += " ORDER BY " + compiler.Content
		}

		return sql, nil
	})

}
