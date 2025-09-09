package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/model"
)

func (selectors *selectorTypes) findByJoin(item any) error {

	modelType := reflect.TypeOf(item)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Slice {
		return fmt.Errorf("%s is not slice", reflect.ValueOf(item).String())
	}
	modelType = modelType.Elem()
	ent, err := model.ModelRegister.GetModelByType(modelType)
	if err != nil {
		return err
	}
	strJoin := ent.Entity.TableName + " " + selectors.strJoin
	strFilter, argsWhere := selectors.getFilter()

	if len(selectors.selectFields) == 0 {
		for _, x := range ent.Entity.Cols {
			selectors.selectFields = append(selectors.selectFields, fmt.Sprintf("%s.%s %s", ent.Entity.TableName, x.Name, x.Field.Name))
		}
	}
	sqlInfo := &types.SqlInfo{
		From:       strJoin,
		StrWhere:   strFilter,
		StrSelect:  strings.Join(selectors.selectFields, ","),
		StrGroupBy: selectors.strGroup,
	}
	sqlQuery, err := compiler.GetSql(sqlInfo, selectors.db.DriverName)
	if err != nil {
		return err
	}
	return selectors.db.fecthItems(item, sqlQuery, nil, nil, true, argsWhere...)
}
