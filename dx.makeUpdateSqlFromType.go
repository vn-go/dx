package dx

import (
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/model"
)

type initMakeUpdateSqlFromType struct {
	once sync.Once
	val  *initMakeUpdateSqlFromTypeItem
	err  error
}
type initMakeUpdateSqlFromTypeItem struct {
	sql           string
	fieldIndex    [][]int
	keyFieldIndex [][]int
}

var makeUpdateSqlFromTypeWithCacheData = sync.Map{}

func makeUpdateSqlFromTypeWithCache(db *DB, typ reflect.Type) (*initMakeUpdateSqlFromTypeItem, error) {
	key := db.Info.DriverName + ":" + typ.String()
	actual, _ := makeUpdateSqlFromTypeWithCacheData.LoadOrStore(key, &initMakeUpdateSqlFromType{})
	init := actual.(*initMakeUpdateSqlFromType)
	init.once.Do(func() {
		init.val, init.err = makeUpdateSqlFromType(db, typ)
	})
	return init.val, init.err

}
func makeUpdateSqlFromType(db *DB, typ reflect.Type) (*initMakeUpdateSqlFromTypeItem, error) {
	ret := initMakeUpdateSqlFromTypeItem{
		sql:           "",
		fieldIndex:    nil,
		keyFieldIndex: nil,
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
				conditional = append(conditional, dialect.Quote(col.Name)+" = "+dialect.ToParam(i+1))
				ret.keyFieldIndex = append(ret.keyFieldIndex, col.IndexOfField)

			}
			continue

		}
		strPlaceHoldesr = append(strPlaceHoldesr, col.Name+" = "+dialect.ToParam(i+1))
		ret.fieldIndex = append(ret.fieldIndex, col.IndexOfField)

	}
	sql += strings.Join(strPlaceHoldesr, ",")
	if len(conditional) > 0 {
		sql += " WHERE " + strings.Join(conditional, " AND ")
	}
	ret.sql = sql
	return &ret, nil
}
