package dx

import (
	"reflect"

	"github.com/vn-go/dx/db"
	dbutils "github.com/vn-go/dx/dbUtils"
	"github.com/vn-go/dx/migate/migrator"
	"github.com/vn-go/dx/model"
)

var modelRegistry = model.ModelRegister

func AddModels(models ...any) {
	for _, model := range models {
		typ := reflect.TypeOf(model)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		modelRegistry.RegisterType(typ)

	}

}

type FkOpt struct {
	OnDelete bool
	OnUpdate bool
}

func AddForeignKey[T any](foreignKey string, FkEntity interface{}, keys string, cascadeOption *FkOpt) error {
	if cascadeOption == nil {
		return model.AddForeignKey[T](foreignKey, FkEntity, keys, nil)
	} else {
		err := model.AddForeignKey[T](foreignKey, FkEntity, keys, &model.CascadeOption{
			OnDelete: cascadeOption.OnDelete,
			OnUpdate: cascadeOption.OnUpdate,
		})
		return err
	}

}
func Open(driverName string, dsn string) (*DB, error) {
	retDb, err := db.Open(driverName, dsn)

	if err != nil {
		return nil, err
	}
	m, err := migrator.GetMigator(retDb)
	if err != nil {
		defer retDb.Close()
		return nil, err
	}
	err = m.DoMigrates(retDb)
	if err != nil {
		defer retDb.Close()
		return nil, err
	}
	return &DB{
		DB: retDb,
	}, nil

}
func SetManagerDb(driver, dbName string) {
	db.SetManagerDb(driver, dbName)
}
func NewDTO[T any]() (*T, error) {
	typ := reflect.TypeFor[T]()
	valOfModel := reflect.New(typ)
	val, err := dbutils.DbUtils.ModelFactory.SetDefaultValue(valOfModel)
	if err != nil {
		return nil, err
	}
	return val.(*T), nil

}
func NewThenSetDefaultValues[T any](fn func() (*T, error)) (*T, error) {

	retErr := reflect.ValueOf(fn).Call([]reflect.Value{})
	if retErr[1].Interface() != nil {
		return nil, retErr[0].Interface().(error)
	}

	val, err := dbutils.DbUtils.ModelFactory.SetDefaultValue(retErr[0])
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return val.(*T), nil

}
func Prt[T any](val T) *T {
	return &val
}
