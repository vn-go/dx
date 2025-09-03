package tenantDB

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type modelRegistryInfo struct {
	tableName string
	modelType reflect.Type
	entity    Entity
}
type modelRegister struct {
	cacheModelRegistry  sync.Map
	cacheGetModelByType sync.Map
}

func (info *modelRegistryInfo) GetColumns() []ColumnDef {
	return info.entity.cols
}
func (info *modelRegistryInfo) GetPrimaryConstraints() map[string][]ColumnDef {
	return info.entity.primaryConstraints
}
func (info *modelRegistryInfo) GetUniqueConstraints() map[string][]ColumnDef {
	return info.entity.uniqueConstraints
}
func (info *modelRegistryInfo) GetIndexConstraints() map[string][]ColumnDef {
	return info.entity.indexConstraints
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
func (reg *modelRegister) getTableNameNoCache(typ reflect.Type) (string, error) {
	// scan field
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Type.Field(0).Type == reflect.TypeOf(Entity{}) {
			tagInfo := field.Tag.Get("db")
			if strings.HasPrefix(tagInfo, "table:") {
				return tagInfo[6:], nil
			}
			return utilsInstance.Pluralize(utilsInstance.SnakeCase(typ.Name())), nil

		}
	}
	return "", fmt.Errorf("model %s has no table tag", typ.String())
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
		entityType:  typ,
		tableName:   tableName,
		DbTableName: tableName,
		cols:        cols,
	}

	entity.primaryConstraints = utilsInstance.GetPrimaryKey(&entity)
	entity.uniqueConstraints = make(map[string][]ColumnDef)
	entity.indexConstraints = make(map[string][]ColumnDef)

	cacheItem := modelRegistryInfo{
		tableName: tableName,
		modelType: typ,
		entity:    entity,
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

		if m.modelType == typ {
			ret = m
			return false // <-- dừng duyệt nếu tìm thấy (RETURN FALSE mới là dừng)
		}
		return true // tiếp tục duyệt
	})

	return ret
}

var ModelRegistry = &modelRegister{}

func (m *modelRegistryInfo) GetTableName() string {
	return m.tableName
}
func (m *modelRegistryInfo) GetEntity() Entity {
	return m.entity

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
	Columns   []ColumnDef
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
		uk := model.entity.getBuildUniqueConstraints()
		if _, ok := uk[name]; ok {
			ret := UKConstraintResult{
				TableName: model.tableName,
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
