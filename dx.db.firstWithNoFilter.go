package dx

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/vn-go/dx/internal"
)

func (db *DB) firstWithNoFilter(entity interface{}, ctx context.Context, sqlTx *sql.Tx) error {
	err := internal.Helper.AddrssertSinglePointerToStruct(entity)
	if err != nil {
		return err
	}
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, _, _, err := db.buildBasicSqlFirstItemNoFilter(typ) //OnBuildSQLFirstItemNoFilter(typ, db)

	if err != nil {
		return err
	}
	return db.ExecToItem(entity, sql, ctx, sqlTx)

}
func (db *DB) firstWithNoFilterV2(entity interface{}, ctx context.Context, sqlTx *sql.Tx) error {
	err := internal.Helper.AddrssertSinglePointerToStruct(entity)
	if err != nil {
		return err
	}
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, _, _, err := db.buildBasicSqlFirstItemNoFilterV2(typ) //OnBuildSQLFirstItemNoFilter(typ, db)

	if err != nil {
		return err
	}
	return db.ExecToItem(entity, sql, ctx, sqlTx)

}
