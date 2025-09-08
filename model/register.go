package model

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/vn-go/dx/entity"
)

var ModelRegister = &modelRegister{
	cacheModelRegistry:      map[reflect.Type]*modelRegistryInfo{},
	cacheGetModelByType:     sync.Map{},
	cacheTableNameAndEntity: map[string]*entity.Entity{},
}

type initGetModelByType struct {
	val  *modelRegistryInfo
	err  error
	once sync.Once
}

func (r *modelRegister) GetModelByType(typ reflect.Type) (*modelRegistryInfo, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	actually, _ := r.cacheGetModelByType.LoadOrStore(typ, &initGetModelByType{})
	item := actually.(*initGetModelByType)
	item.once.Do(func() {
		if ret, ok := ModelRegister.cacheModelRegistry[typ]; ok {
			ret.Entity.IndexConstraints = ret.Entity.GetIndex()
			ret.Entity.PrimaryConstraints = ret.Entity.GetPrimaryKey()
			ret.Entity.UniqueConstraints = ret.Entity.GetUnique()

			item.val = ret
		} else {
			item.err = fmt.Errorf("%s was not register", typ.String())
		}
	})
	return item.val, item.err

}
func GetModel[T any]() (*modelRegistryInfo, error) {
	return ModelRegister.GetModelByType(reflect.TypeFor[T]())
}

var onceGetAllModels sync.Once
var allModels = make([]*modelRegistryInfo, 0)

func (reg *modelRegister) GetAllModels() []*modelRegistryInfo {

	onceGetAllModels.Do(func() {
		for _, r := range reg.cacheModelRegistry {
			allModels = append(allModels, r)
		}
	})

	return allModels
}
