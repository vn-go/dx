package errors

import (
	"fmt"
	"reflect"

	migrate "github.com/vn-go/dx/migrate/mssql"
	// "vdb/migrate"
)

type ModelError struct {
	typ reflect.Type
}

func (e *ModelError) Error() string {
	typ := e.typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return fmt.Sprintf("%s is not recognized as a model or has not been registered,Please embed vdb.Model[%s] in %s struct and call vdb.ModelRegistry.Add(%s)", typ.String(), typ.Name(), typ.Name(), typ.Name())
}
func NewModelError(typ reflect.Type) error {
	return &ModelError{typ: typ}
}
func init() {
	migrate.NewModelError = NewModelError

}

type UnSupportError struct {
	msg string
}

func (e *UnSupportError) Error() string {
	return e.msg
}
func NewUnSupportError(msg string) error {
	return &UnSupportError{msg: msg}

}

type InvalidDefaultValue struct {
	field        reflect.StructField
	eleType      reflect.Type
	defaultValue string
}

func (e *InvalidDefaultValue) Error() string {
	return fmt.Sprintf("type of %s.%s is %s does not fix %s", e.eleType.String(), e.field.Name, e.field.Type.String(), e.defaultValue)
}
func NewInvalidDefaultValue(field reflect.StructField, eleType reflect.Type, defaultValue string) error {
	return &InvalidDefaultValue{field: field, eleType: eleType, defaultValue: defaultValue}
}
