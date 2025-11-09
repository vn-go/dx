package dx

import (
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
	"github.com/vn-go/dx/sqlparser"
)

type initMakeUpdateSqlFromTypeItem struct {
	sql              string
	fieldIndex       [][]int
	keyFieldIndex    [][]int
	keyFieldIndexPos []int
}

func makeUpdateSqlFromTypeWithCache(db *DB, typ reflect.Type) (*initMakeUpdateSqlFromTypeItem, error) {
	key := db.DbName + "@" + db.DriverName + "@" + typ.String()
	return internal.OnceCall(key, func() (*initMakeUpdateSqlFromTypeItem, error) {
		ret := initMakeUpdateSqlFromTypeItem{
			sql:              "",
			fieldIndex:       nil,
			keyFieldIndex:    nil,
			keyFieldIndexPos: []int{},
		}

		model, err := model.ModelRegister.GetModelByType(typ)
		if err != nil {
			return nil, err
		}

		dialect := factory.DialectFactory.Create(db.DriverName)

		sql := "UPDATE " + dialect.Quote(model.Entity.TableName) + " SET "
		conditional := []string{}

		strPlaceHoldesr := []string{}

		for i, col := range model.Entity.Cols {
			if col.PKName != "" || col.IsAuto {
				if col.PKName != "" {
					conditional = append(conditional, dialect.Quote(col.Name)+" = "+dialect.ToParam(i+1, sqlparser.ValArg))
					ret.keyFieldIndex = append(ret.keyFieldIndex, col.IndexOfField)
					ret.keyFieldIndexPos = append(ret.keyFieldIndexPos, i)
					ret.fieldIndex = append(ret.fieldIndex, col.IndexOfField)

				}
				continue

			}
			strPlaceHoldesr = append(strPlaceHoldesr, col.Name+" = "+dialect.ToParam(i+1, sqlparser.ValArg))
			ret.fieldIndex = append(ret.fieldIndex, col.IndexOfField)
			ret.keyFieldIndexPos = append(ret.keyFieldIndexPos, i)
		}
		sql += strings.Join(strPlaceHoldesr, ",")
		if len(conditional) > 0 {
			sql += " WHERE " + strings.Join(conditional, " AND ")
		}
		ret.sql = sql
		return &ret, nil
	})

}
