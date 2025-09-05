package dx

import (
	dbutils "github.com/vn-go/dx/dbUtils"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/migate/migrator"
)

func (db *DB) Insert(data interface{}) error {
	err := internal.Helper.AddrssertSinglePointerToStruct(data)
	if err != nil {
		return err
	}

	m, err := migrator.GetMigator(db.DB)
	if err != nil {
		return err
	}
	err = m.DoMigrates()
	if err != nil {
		return err
	}

	return dbutils.DbUtils.Insert.Insert(db.DB, data)
}
