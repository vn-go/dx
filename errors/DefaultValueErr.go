package errors

import (
	"fmt"
	"reflect"
)

type DFErr struct {
	field        reflect.StructField
	eleType      reflect.Type
	defaultValue string
}

func (e *DFErr) Error() string {
	return fmt.Sprintf("type of %s.%s is %s does not fix %s", e.eleType.String(), e.field.Name, e.field.Type.String(), e.defaultValue)
}
func NewDFErr(field reflect.StructField, eleType reflect.Type, defaultValue string) error {
	return &DFErr{field: field, eleType: eleType, defaultValue: defaultValue}
}
