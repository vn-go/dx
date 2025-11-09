package dx

import (
	"context"
	"fmt"
	"reflect"

	"github.com/vn-go/dx/internal"
)

func (db *DB) UpdateWithContext(context context.Context, item interface{}) UpdateResult {
	val := reflect.ValueOf(item).Elem()
	typ := reflect.TypeOf(item)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()

	}
	info, err := makeUpdateSqlFromTypeWithCache(db, typ)
	if err != nil {
		return UpdateResult{RowsAffected: 0, Sql: "", Error: err}
	}

	args := make([]interface{}, len(info.fieldIndex)+len(info.keyFieldIndex))
	for i, index := range info.fieldIndex {
		args[i] = val.FieldByIndex(index).Interface()
	}
	//numOfFieds := len(info.fieldIndex)
	for i, index := range info.keyFieldIndex {
		//keyField := typ.FieldByIndex(index)

		args[info.keyFieldIndexPos[i]] = val.FieldByIndex(index).Interface()
	}
	if db.DriverName == "mysql" {
		info.sql, args, err = internal.Helper.FixParam(info.sql, args)
		if err != nil {
			return UpdateResult{
				Error: err,
			}
		}
	}
	if Options.ShowSql {
		fmt.Println("---------------------")
		fmt.Println(info.sql)
		fmt.Println("---------------------")
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
func (db *DB) Update(item interface{}) UpdateResult {

	return db.UpdateWithContext(context.Background(), item)
}
