package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/expr"
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

	compiler, err := expr.CompileJoin(strJoin, selectors.db.DB)
	if err != nil {
		return err
	}
	compilerTableName := []string{}
	compilerAlias := map[string]string{}
	aliasMap := map[string]string{}
	
	for _, tblName := range compiler.Context.Tables {
		if x := model.ModelRegister.FindEntityByName(tblName); x != nil {
			if _, ok := compilerAlias[x.TableName]; !ok {
				compilerTableName = append(compilerTableName, x.TableName)
				alias := compiler.Context.Alias[tblName]
				compilerAlias[tblName] = x.TableName
				if _, ok2 := compiler.Context.Alias[x.TableName]; ok2 {
					compiler.Context.Alias[x.TableName] = alias
				}
				aliasMap[x.TableName] = alias
			} else {
				fmt.Println(tblName)
			}
		}
	}
	compiler.Context.AlterTableJoin = compilerAlias
	compiler.Context.Tables = compilerTableName
	//compiler.Context.Alias = aliasMap
	compiler.Build(strJoin)
	selectors.strJoin = compiler.Content
	strWhere, whereArgs := selectors.getFilter()
	if len(selectors.selectFields) == 0 {
		for _, x := range ent.Entity.Cols {
			selectors.selectFields = append(selectors.selectFields, fmt.Sprintf("%s.%s %s", ent.Entity.TableName, x.Name, x.Field.Name))
		}
	}
	strSelect := strings.Join(selectors.selectFields, ",")

	err = compiler.BuildSelectField(strSelect)
	if err != nil {
		return err
	}
	strSelect = compiler.Content
	if strWhere != "" {
		err = compiler.BuildWhere(strWhere)
		if err != nil {
			return err
		}
		strWhere = compiler.Content
	}
	sqlInfo := &types.SqlInfo{
		From:      selectors.strJoin,
		StrWhere:  strWhere,
		StrSelect: strSelect,
	}
	sqlQuery, err := compiler.Context.Dialect.BuildSql(sqlInfo)
	if err != nil {
		return err
	}
	return selectors.db.fecthItems(item, sqlQuery, nil, nil, true, whereArgs...)
}
