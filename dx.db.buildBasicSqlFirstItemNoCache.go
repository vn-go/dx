package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/model"
)

func (db *DB) buildBasicSqlFirstItemNoCache(typ reflect.Type, filter string) (string, error) {
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
	sql = dialect.MakeSelectTop(sql, 1)
	return sql, nil
}
