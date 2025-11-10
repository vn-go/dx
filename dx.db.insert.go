package dx

import (
	"context"
	"reflect"

	dbutils "github.com/vn-go/dx/dbUtils"
	"github.com/vn-go/dx/internal"
)

func (db *DB) Insert(data interface{}) error {
	err := internal.Helper.AddrssertSinglePointerToStruct(data)
	if err != nil {
		return err
	}
	reflect.ValueOf(data).Field(0).Interface()
	return dbutils.DbUtils.Insert.Insert(db.DB, data, context.Background(), Options.ShowSql)
}
