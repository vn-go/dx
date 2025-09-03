package common

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/vn-go/dx/internal"
)

type modelRegistryInfo struct {
	TableName string
	ModelType reflect.Type
	Entity    *Entity
}
type modelRegister struct {
	cacheModelRegistry  sync.Map
	cacheGetModelByType sync.Map
}

func (info *modelRegistryInfo) GetColumns() []internal.ColumnDef {
	return info.Entity.Cols
}
func (info *modelRegistryInfo) GetPrimaryConstraints() map[string][]internal.ColumnDef {
	return info.Entity.PrimaryConstraints
}
func (info *modelRegistryInfo) GetUniqueConstraints() map[string][]internal.ColumnDef {
	return info.Entity.uniqueConstraints
}
func (info *modelRegistryInfo) GetIndexConstraints() map[string][]internal.ColumnDef {
	return info.Entity.indexConstraints
}

type initGetTableName struct {
	once sync.Once
	val  string
	err  error
}

var cacheGetTableName sync.Map

func (reg *modelRegister) getTableName(typ reflect.Type) (string, error) {
	actual, _ := cacheGetTableName.LoadOrStore(typ, &initGetTableName{})
	init := actual.(*initGetTableName)
	init.once.Do(func() {
		init.val, init.err = reg.getTableNameNoCache(typ)
	})
	return init.val, init.err
}
func (reg *modelRegister) getTableNameFromTableFunc(typ reflect.Type) string {
	if typ.Kind() == reflect.Struct {
		typ = reflect.PointerTo(typ)
	}
	for i := 0; i < typ.NumMethod(); i++ {
		if typ.Method(i).Name == "Table" {
			ret := typ.Method(i).Func.Call([]reflect.Value{reflect.ValueOf(typ.Elem())})
			if len(ret) == 1 && ret[0].Type() == reflect.TypeFor[string]() {
				return ret[0].String()
			}
		}
	}
	return ""
}
func (reg *modelRegister) getTableNameNoCache(typ reflect.Type) (string, error) {
	// scan field
	if ret := reg.getTableNameFromTableFunc(typ); ret != "" {
		return ret, nil
	}
	return utilsInstance.Pluralize(utilsInstance.SnakeCase(typ.Name())), nil

}
func (reg *modelRegister) RegisterType(typ reflect.Type) {
	if _, ok := reg.cacheModelRegistry.Load(typ); ok {
		return
	}
	tableName, err := reg.getTableName(typ)

	if err != nil {
		fmt.Println(typ.String())
		panic(err)
	}
	cols, err := utilsInstance.ParseStruct(typ, []int{})
	if err != nil {
		panic(err)
	}
	entity := Entity{
		EntityType:  typ,
		tableName:   tableName,
		DbTableName: tableName,
		Cols:        cols,
	}

	entity.PrimaryConstraints = utilsInstance.GetPrimaryKey(&entity)
	entity.uniqueConstraints = make(map[string][]internal.ColumnDef)
	entity.indexConstraints = make(map[string][]internal.ColumnDef)

	cacheItem := modelRegistryInfo{
		TableName: tableName,
		ModelType: typ,
		Entity:    &entity,
	}

	reg.cacheModelRegistry.Store(typ, &cacheItem)
}

func (reg *modelRegister) Add(m ...interface{}) {
	for _, model := range m {

		typ := reflect.TypeOf(model)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		reg.RegisterType(typ)
	}
}
func (reg *modelRegister) GetAllModels() []*modelRegistryInfo {
	ret := make([]*modelRegistryInfo, 0)
	reg.cacheModelRegistry.Range(func(key, value interface{}) bool {
		if v, ok := value.(*modelRegistryInfo); ok {
			ret = append(ret, v)
		}
		return true
	})
	return ret
}

type getModelByTypeInit struct {
	once sync.Once
	ret  *modelRegistryInfo
}

func (reg *modelRegister) GetModelByType(typ reflect.Type) *modelRegistryInfo {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	actual, _ := reg.cacheGetModelByType.LoadOrStore(typ, &getModelByTypeInit{})
	init := actual.(*getModelByTypeInit)
	init.once.Do(func() {
		init.ret = reg.getModelByType(typ)
	})
	return init.ret
}
func (reg *modelRegister) getModelByType(typ reflect.Type) *modelRegistryInfo {

	var ret *modelRegistryInfo
	reg.cacheModelRegistry.Range(func(key, value interface{}) bool {
		m := value.(*modelRegistryInfo)

		if m.ModelType == typ {
			ret = m
			return false // <-- dừng duyệt nếu tìm thấy (RETURN FALSE mới là dừng)
		}
		return true // tiếp tục duyệt
	})

	return ret
}

var ModelRegistry = &modelRegister{}

func (m *modelRegistryInfo) GetTableName() string {
	return m.TableName
}
func (m *modelRegistryInfo) GetEntity() *Entity {
	return m.Entity

}
func GetModelByType(typ reflect.Type) *modelRegistryInfo {
	return ModelRegistry.GetModelByType(typ)
}

type initFindUKConstraint struct {
	once sync.Once
	val  *UKConstraintResult
}
type UKConstraintResult struct {
	TableName string
	Columns   []internal.ColumnDef
	DbCols    []string
	Fields    []string
	//Columns   []string
}

var cacheFindUKConstraint sync.Map

func FindUKConstraint(name string) *UKConstraintResult {
	actual, _ := cacheFindUKConstraint.LoadOrStore(name, &initFindUKConstraint{})
	init := actual.(*initFindUKConstraint)
	init.once.Do(func() {
		init.val = findUKConstraint(name)
	})
	return init.val
}

func findUKConstraint(name string) *UKConstraintResult {
	for _, model := range ModelRegistry.GetAllModels() {
		uk := model.Entity.getBuildUniqueConstraints()
		if _, ok := uk[name]; ok {
			ret := UKConstraintResult{
				TableName: model.TableName,
				Columns:   uk[name],
			}
			for _, col := range uk[name] {
				ret.DbCols = append(ret.DbCols, col.Name)
				ret.Fields = append(ret.Fields, col.Field.Name)
			}
			return &ret
		}
	}
	return nil
}

type FKConstraintResult struct {
	FormTable  string
	ToTable    string
	FromCols   []string
	ToCols     []string
	FromFields []string
	ToFields   []string
}
