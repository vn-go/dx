package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/model"
)

func (db *DB) buildBasicSqlNoCache(typ reflect.Type, filter string) (string, error) {

	repoType, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return "", err
	}
	tableName := repoType.Entity.TableName
	// compiler, err := expr.CompileJoin(tableName, db.DB)
	// if err != nil {
	// 	return "", err
	// }
	// tableName = compiler.Content
	columns := repoType.Entity.Cols

	fieldsSelect := make([]string, len(columns))
	for i, col := range columns {
		fieldsSelect[i] = col.Name + " AS " + col.Field.Name
	}
	// compiler.Context.Purpose = expr.BUILD_SELECT
	// err = compiler.BuildSelectField(strings.Join(fieldsSelect, ", "))
	// if err != nil {
	// 	return "", err
	// }
	// strField := compiler.Content

	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fieldsSelect, ","), tableName)
	if filter != "" {
		// compiler.Context.Purpose = expr.BUILD_WHERE
		// err = compiler.BuildWhere(filter)
		// if err != nil {
		// 	return "", err
		// }
		sql += " WHERE " + filter
	}
	sqlInfo, err := compiler.Compile(sql, db.DriverName, false, false)
	if err != nil {
		return "", err
	}
	sqlParse, err := factory.DialectFactory.Create(db.DriverName).BuildSql(sqlInfo.Info)
	if err != nil {
		return "", err
	}

	return sqlParse.Sql, nil
}
