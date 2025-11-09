package dx

import (
	"reflect"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/db"
	dbutils "github.com/vn-go/dx/dbUtils"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/errors"
	dxErrors "github.com/vn-go/dx/errors"
	migrateLoaderTypes "github.com/vn-go/dx/migrate/loader/types"
	"github.com/vn-go/dx/migrate/migrator"
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
	err = m.DoMigrates(retDb, m.GetDefaultSchema())
	if err != nil {
		defer retDb.Close()
		return nil, err
	}
	dialect := factory.DialectFactory.Create(retDb.Info.DriverName)

	return &DB{
		DB:      retDb,
		Dialect: dialect,
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
func Ptr[T any](val T) *T {
	return &val
}
func IsNull[T any](val *T, defVal T) T {
	if val != nil {
		return *val
	} else {
		return defVal
	}
}

type dbErrType int

type errorTypes struct {
	UNKNOWN    dbErrType //unknown error
	DUPLICATE  dbErrType // database return error duplicate value
	REFERENCES dbErrType // database return error foreign key constraint value
	REQUIRED   dbErrType // database return error field require value
	SYSTEM     dbErrType // database return error
	NOTFOUND   dbErrType
	ERR_SYNTAX dbErrType
}

/*
type DB_ERR = int

const (

	ERR_UNKNOWN DB_ERR = iota
	ERR_DUPLICATE
	ERR_REFERENCES // ✅ refrences_violation
	ERR_REQUIRED
	ERR_SYSTEM

)
*/
func String(t dbErrType) string {
	return dxErrors.ErrorMessage(dxErrors.DB_ERR(t))

}

var Errors = &errorTypes{
	UNKNOWN:    dbErrType(dxErrors.ERR_UNKNOWN),
	DUPLICATE:  dbErrType(dxErrors.ERR_DUPLICATE),
	REFERENCES: dbErrType(dxErrors.ERR_REFERENCES),
	REQUIRED:   dbErrType(dxErrors.ERR_REQUIRED),
	SYSTEM:     dbErrType(dxErrors.ERR_SYSTEM),
	NOTFOUND:   dbErrType(dxErrors.ERR_NOT_FOUND),
	ERR_SYNTAX: dbErrType(dxErrors.ERR_SYNTAX),
}

type DbError struct {
	dxErrors.DbErr
	ErrorType dbErrType
}

func (d *DbError) IsDuplicateEntryError() bool {
	return d.ErrorType == Errors.DUPLICATE
}

func (er *errorTypes) IsDbError(err error) *DbError {
	if ret, ok := err.(*dxErrors.DbErr); ok {
		retErr := &DbError{
			ErrorType: dbErrType(ret.ErrorType),
			DbErr:     *ret,
		}
		return retErr
	} else if _, ok := err.(*errors.NotFoundErr); ok {
		retErr := &DbError{
			ErrorType: Errors.NOTFOUND,
		}
		return retErr

	}
	return nil
}
func (er *errorTypes) IsExpressionError(err error) *compiler.CompilerError {
	if _, ok := err.(*compiler.CompilerError); ok {
		return err.(*compiler.CompilerError)
	} else {
		return nil
	}
}
func SkipLoadSchemOnMigrate(ok bool) {
	migrateLoaderTypes.SkipLoadSchemaOnMigrate = ok

}
