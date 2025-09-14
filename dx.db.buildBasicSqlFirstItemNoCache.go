package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/model"
)

func (db *DB) buildBasicSqlFirstItemNoCache(typ reflect.Type, filter string) (string, error) {
	dialect := factory.DialectFactory.Create(db.DriverName)

	repoType, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return "", err
	}
	tableName := repoType.Entity.TableName

	columns := repoType.Entity.Cols

	fieldsSelect := make([]string, len(columns))
	for i, col := range columns {
		fieldsSelect[i] = repoType.Entity.TableName + "." + col.Field.Name + " AS " + col.Field.Name
	}

	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fieldsSelect, ","), tableName)
	if filter != "" {

		sql += " WHERE " + filter
	}
	sqlInfo, err := compiler.Compile(sql, db.DriverName)
	if err != nil {
		return "", err
	}
	sqlInfo.Limit = Ptr[uint64](1)
	// offset := uint64(1)
	sqlParse, err := dialect.BuildSql(sqlInfo)
	if err != nil {
		return "", err
	}
	//sql = dialect.LimitAndOffset(sql, nil, &offset, "")
	return sqlParse.Sql, nil
}
