package model

import (
	sysErrors "errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/migrator/types"
)

type CascadeOption struct {
	OnDelete bool
	OnUpdate bool
}

func AddForeignKey[T any](foreignKey string, FkEntity interface{}, keys string, cascadeOption *CascadeOption) error {
	if cascadeOption == nil {
		cascadeOption = &CascadeOption{
			OnDelete: true,
			OnUpdate: true,
		}
	}

	ks := strings.Split(keys, ",")
	fks := strings.Split(foreignKey, ",")
	FkEntityType := reflect.TypeFor[T]()
	ownerType := reflect.TypeOf(FkEntity)
	if ownerType.Kind() == reflect.Ptr {
		ownerType = ownerType.Elem()
	}
	ModelRegister.RegisterType(ownerType)
	ModelRegister.RegisterType(FkEntityType)

	ownerInfo, err := ModelRegister.GetModelByType(ownerType)
	if err != nil {
		return err
	}
	if ownerInfo == nil {
		panic(errors.NewModelError(ownerType))
	}
	fkInfo, err := ModelRegister.GetModelByType(FkEntityType)
	if err != nil {
		return err
	}
	if fkInfo == nil {
		return (errors.NewModelError(FkEntityType))
	}
	if FkEntityType.Kind() == reflect.Ptr {
		FkEntityType = FkEntityType.Elem()
	}

	if len(ks) != len(fks) {
		return (fmt.Errorf("len of key and foreign key not match: %s(%s)!= %s(%s)", ownerInfo.Entity.TableName, keys, fkInfo.Entity.TableName, foreignKey))
	}
	ownerMap := map[string]entity.ColumnDef{}
	for _, col := range ownerInfo.Entity.Cols {
		ownerMap[strings.ToLower(col.Field.Name)] = col
	}
	fkMap := map[string]entity.ColumnDef{}

	for _, col := range fkInfo.Entity.Cols {
		fkMap[strings.ToLower(col.Field.Name)] = col

	}
	pkCols := []string{}
	pkFields := []string{}
	fkColsName := []string{}
	fkFieldsName := []string{}
	for i, key := range ks {
		keyCol := ownerMap[strings.ToLower(key)]
		fkCol := fkMap[strings.ToLower(fks[i])]
		if keyCol.Type != fkCol.Type {
			errText := fmt.Sprintf("foreign key column not match with primary key of %s.%s and %s.%s ", ownerType.String(), keyCol.Name, FkEntityType.String(), fkCol.Name)
			panic(sysErrors.New(errText))
		}
		pkCols = append(pkCols, keyCol.Name)
		pkFields = append(pkFields, keyCol.Field.Name)
		fkColsName = append(fkColsName, fkCol.Name)
		fkFieldsName = append(fkFieldsName, fkCol.Field.Name)

	}

	types.ForeignKeyRegistry.Register(&types.ForeignKeyInfo{
		FromTable:      fkInfo.Entity.TableName,
		FromCols:       fkColsName,
		ToTable:        ownerInfo.Entity.TableName,
		ToCols:         pkCols,
		FromStructName: FkEntityType.String(),
		ToStructName:   ownerType.String(),
		FromFiels:      fkFieldsName,
		ToFiels:        pkFields,
	})

	return nil

}
