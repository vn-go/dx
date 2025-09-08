package model

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/entity"
)

type initRegisterType struct {
	val  entity.Entity
	once sync.Once
}

var cacheModelRegistry sync.Map

func (reg *modelRegister) RegisterType(typ reflect.Type) {
	actual, _ := cacheModelRegistry.LoadOrStore(typ, &initRegisterType{})
	item := actual.(*initRegisterType)
	item.once.Do(func() {
		// if _, ok := cacheModelRegistry.Load(typ); ok {
		// 	return
		// }
		tableName, err := reg.getTableName(typ)

		if err != nil {
			fmt.Println(typ.String())
			panic(err)
		}
		cols, err := entity.Reader.ParseStruct(typ, []int{})
		if err != nil {
			panic(err)
		}
		ent := entity.Entity{
			EntityType: typ,
			TableName:  tableName,
			//DbTableName: tableName,
			Cols: cols,
		}
		ent.PrimaryConstraints = ent.GetPrimaryKey()

		ent.UniqueConstraints = entity.NewUniqueConstraints()
		ent.IndexConstraints = make(map[string][]entity.ColumnDef)

		cacheItem := modelRegistryInfo{

			ModelType: typ,
			Entity:    &ent,
		}

		reg.cacheModelRegistry[typ] = &cacheItem
		reg.cacheTableNameAndEntity[strings.ToLower(typ.Name())] = &ent
	})

}
func (reg *modelRegister) FindEntityByName(name string) *entity.Entity {
	if ret, ok := reg.cacheTableNameAndEntity[strings.ToLower(name)]; ok {
		return ret
	}
	return nil
}
func (reg *modelRegister) GetMapEntities(tables []string) map[string]*entity.Entity {
	ret := map[string]*entity.Entity{}
	for _, tblName := range tables {
		if x := reg.FindEntityByName(tblName); x != nil {
			ret[strings.ToLower(tblName)] = x
		}
	}
	return ret
}
