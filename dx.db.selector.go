package dx

import (
	"reflect"
)

func (dbCtx *dbContext) Model(ent any) *modelType {
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
		db:         dbCtx.DB,
		typ:        typ,
		typEle:     typeEle,
		valuaOfEnt: reflect.ValueOf(ent).Elem(),
		ctx:        dbCtx.ctx,
	}
}
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
func (tx *Tx) Model(ent any) *modelType {
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
		db:         tx.db,
		typ:        typ,
		typEle:     typeEle,
		valuaOfEnt: reflect.ValueOf(ent).Elem(),
		tx:         tx,
	}
}
