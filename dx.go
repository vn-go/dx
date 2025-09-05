package dx

import (
	"reflect"

	"github.com/vn-go/dx/db"
	dbutils "github.com/vn-go/dx/dbUtils"
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
	return nil
}
func Open(driverName string, dsn string) (*DB, error) {
	rdbret, err := db.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	return &DB{
		DB: rdbret,
	}, nil

}
func SetManagerDb(driver, dbName string) {
	db.SetManagerDb(driver, dbName)
}
func NewDTO[T any]() (*T, error) {
	typ := reflect.TypeFor[T]()
	val, err := dbutils.DbUtils.ModelFactory.CreateFromType(typ)
	if err != nil {
		return nil, err
	}
	return val.(*T), nil

}
