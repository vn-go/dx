package mysql

import (
	"github.com/go-sql-driver/mysql"

	"github.com/vn-go/dx/errors"
)

func (d *mySqlDialect) ParseError1064(err *mysql.MySQLError) *errors.DbErr {

	return &errors.DbErr{
		ErrorType: errors.ERR_SYNTAX,
		Code:      "Error syntax",
	}

}
