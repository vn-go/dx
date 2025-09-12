package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/model"
)

func (cmp *compiler) createDictionary(tables []string) *Dictionary {
	tableAlias := map[string]string{}
	tblList := []string{}
	i := 1
	manualAlaisMap := map[string]string{}
	for _, x := range tables {
		items := strings.Split(x, "\n")
		if len(items) > 1 {
			manualAlaisMap[strings.ToLower(items[0])] = items[1]
			tblList = append(tblList, items[0])
		} else {
			tableAlias[strings.ToLower(x)] = fmt.Sprintf("T%d", i)
			tblList = append(tblList, x)
			i++
		}
	}
	mapEntities := model.ModelRegister.GetMapEntities(tblList)
	ret := &Dictionary{
		TableAlias:  map[string]string{},
		Field:       map[string]string{},
		StructField: map[string]reflect.StructField{},
		Tables:      tables,
	}
	ret.TableAlias = tableAlias
	// mapEntityTypes := map[reflect.Type]string{}
	// count := 1
	newMap := map[string]string{}
	//mapAlias := map[string]string{}
	typeToAlias := map[reflect.Type]string{}
	c := 1
	moreMapEntity := map[string]*entity.Entity{}
	for tbl, x := range mapEntities {
		if mAlias, ok := manualAlaisMap[tbl]; ok {
			newMap[tbl] = mAlias
			typeToAlias[x.EntityType] = mAlias
			moreMapEntity[strings.ToLower(mAlias)] = x

		} else {
			alais, foud := typeToAlias[x.EntityType]
			if !foud {
				typeToAlias[x.EntityType] = fmt.Sprintf("T%d", c)
				newMap[tbl] = fmt.Sprintf("T%d", c)
				c++
			} else {
				newMap[tbl] = alais
			}
		}
	}
	for k, v := range moreMapEntity {
		mapEntities[k] = v
	}
	for tbl, x := range mapEntities {
		alias := typeToAlias[x.EntityType]
		for _, col := range x.Cols {

			key := strings.ToLower(fmt.Sprintf("%s.%s", tbl, col.Field.Name))
			ret.Field[key] = cmp.dialect.Quote(alias, col.Name)
			ret.StructField[key] = col.Field

		}
	}

	ret.TableAlias = newMap
	return ret
}
