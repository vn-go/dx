package dx

import (
	"context"
	"reflect"
)

type modelType struct {
	db         *DB
	typ        reflect.Type
	typEle     reflect.Type
	valuaOfEnt reflect.Value
	ctx        context.Context
	tx         *Tx
}

// type modelTypeSelect struct {
// 	modelType
// 	fields []string
// }
