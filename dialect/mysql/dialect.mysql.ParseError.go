package mysql

import (
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/vn-go/dx/migate/loader/types"
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
		fmt.Println(mysqlErr.Number)

	}
	return err
}
