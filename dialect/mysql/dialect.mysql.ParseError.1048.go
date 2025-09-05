package mysql

import (
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/vn-go/dx/errors"
)

func (d *mySqlDialect) ParseError1048(err *mysql.MySQLError) *errors.DbErr {
	col := ""
	if strings.Contains(err.Message, "Column '") {
		col = strings.Split(err.Message, "'")[1]
		col = strings.Split(col, "'")[0]
	}
	ret := &errors.DbErr{
		Err:       err,
		ErrorType: errors.ERR_REQUIRED,
		DbCols:    []string{col},

		ErrorMessage: "require",
	}
	ret.Reload()
	return ret
}
