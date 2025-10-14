package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/vn-go/dx/migrate/loader/types"
)

func (d *mySqlDialect) ParseError(dbSchema *types.DbSchema, err error) error {
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		if mysqlErr.Number == 1048 { //1452
			return d.ParseError1048(mysqlErr)

		}
		if mysqlErr.Number == 1062 {
			return d.ParseError1062(mysqlErr)

		}
		if mysqlErr.Number == 1452 {
			return d.ParseError1452(mysqlErr)

		}
		if mysqlErr.Number == 1064 {
			return d.ParseError1064(mysqlErr)
		}

	}
	return err
}
