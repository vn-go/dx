package dx

import (
	"reflect"
)

func (db *DB) Model(ent any) *modelType {
	typ := reflect.TypeOf(ent)
	typeEle := typ
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Slice {
		typeEle = typeEle.Elem()
	}
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}
	return &modelType{
		db:         db,
		typ:        typ,
		typEle:     typeEle,
		valuaOfEnt: reflect.ValueOf(ent).Elem(),
	}
}
