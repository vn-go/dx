package dx

import (
	"context"
	"reflect"
)

func (db *DB) UpdateWithContext(context context.Context, item interface{}) UpdateResult {
	typ := reflect.TypeOf(item)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()

	}
	info, err := makeUpdateSqlFromTypeWithCache(db, typ)
	if err != nil {
		return UpdateResult{RowsAffected: 0, Sql: "", Error: err}
	}
	val := reflect.ValueOf(item).Elem()
	args := make([]interface{}, 0)
	for _, index := range info.fieldIndex {
		args = append(args, val.FieldByIndex(index).Interface())
	}
	for _, index := range info.keyFieldIndex {
		args = append(args, val.FieldByIndex(index).Interface())
	}
	r, err := db.ExecContext(context, info.sql, args...)
	if err != nil {
		return UpdateResult{RowsAffected: 0, Sql: info.sql, Error: err}
	}
	n, err := r.RowsAffected()
	if err != nil {
		return UpdateResult{RowsAffected: 0, Sql: info.sql, Error: err}
	}
	return UpdateResult{RowsAffected: n, Sql: info.sql, Error: nil}

}
