package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

type findByJoinKey struct {
	modelType    reflect.Type
	selectorsKey string
}

func (selectors *selectorTypes) findByJoin(item any) error {
	modelType := selectors.GetModelType()
	if modelType == nil {
		mType := reflect.TypeOf(item)
		if mType.Kind() == reflect.Ptr {
			mType = mType.Elem()
		}
		if mType.Kind() != reflect.Slice {
			return fmt.Errorf("%s is not slice", reflect.ValueOf(item).String())
		}
		mType = mType.Elem()
		modelType = &mType
	}

	ent, err := model.ModelRegister.GetModelByType(*modelType)
	if err != nil {
		return err
	}
	selectors.strJoin = ent.Entity.TableName + " " + selectors.strJoin
	key := findByJoinKey{
		modelType:    *modelType,
		selectorsKey: selectors.getKey(),
	}
	// key := modelType.String() + "://selectorTypes/findByJoin/" + selectors.getKey()
	sqlQuery, err := internal.OnceCall(key, func() (*types.SqlParse, error) {
		if len(selectors.selectFields) == 0 {
			for _, x := range ent.Entity.Cols {
				selectors.selectFields = append(selectors.selectFields, fmt.Sprintf("%s.%s %s", ent.Entity.TableName, x.Name, x.Field.Name))
			}
		}
		strSelect := strings.Join(selectors.selectFields, ",")
		sqlInfo := &types.SqlInfo{
			From:       selectors.strJoin,
			StrWhere:   selectors.strWhere,
			StrSelect:  strSelect,
			StrGroupBy: selectors.strGroup,
			FieldArs:   *selectors.args.GetFields(),
			Limit:      selectors.limit,
			Offset:     selectors.offset,
			StrHaving:  selectors.strHaving,
			StrOrder:   selectors.strSort,
		}
		return compiler.GetSql(sqlInfo, selectors.db.DriverName)

	})
	if err != nil {
		return err
	}
	args := selectors.args.GetArgs(sqlQuery.ArgIndex)
	if Options.ShowSql {
		fmt.Println(sqlQuery.Sql)
	}
	return selectors.db.fecthItems(item, sqlQuery.Sql, nil, nil, true, args...)
}
