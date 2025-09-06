package dx

import (
	"reflect"
)

type modelType struct {
	db         *DB
	typ        reflect.Type
	typEle     reflect.Type
	valuaOfEnt reflect.Value
}
type modelTypeSelect struct {
	modelType
	fields []string
}
