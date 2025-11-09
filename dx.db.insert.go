package dx

import (
	"context"

	dbutils "github.com/vn-go/dx/dbUtils"
	"github.com/vn-go/dx/internal"
)

func (db *DB) Insert(data interface{}) error {
	err := internal.Helper.AddrssertSinglePointerToStruct(data)
	if err != nil {
		return err
	}

	return dbutils.DbUtils.Insert.Insert(db.DB, data, context.Background(), Options.ShowSql)
}
