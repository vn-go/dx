package mysql

import (
	"regexp"

	"github.com/go-sql-driver/mysql"
	"github.com/vn-go/dx/errors"
)

func (d *mySqlDialect) ParseError1452(err *mysql.MySQLError) *errors.DbErr {

	re := regexp.MustCompile("CONSTRAINT `([^`]+)`")
	match := re.FindStringSubmatch(err.Message)
	if len(match) > 1 {

		ret := &errors.DbErr{
			ErrorType:      errors.ERR_REFERENCES,
			ConstraintName: match[1],
		}
		ret.Reload()
		return ret
	} else {
		return nil
	}

	return nil

}
