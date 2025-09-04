package errors

import (
	"fmt"
	"reflect"
)

type ModelError struct {
	typ reflect.Type
}

func (e *ModelError) Error() string {
	typ := e.typ
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return fmt.Sprintf("%s is not recognized as a model or has not been registered,Please embed xdb.Model[%s] in %s struct and call xdb.ModelRegistry.Add(%s)", typ.String(), typ.Name(), typ.Name(), typ.Name())
}
func NewModelError(typ reflect.Type) error {
	return &ModelError{typ: typ}
}
