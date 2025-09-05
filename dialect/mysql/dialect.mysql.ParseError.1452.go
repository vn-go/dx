package mysql

import (
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/vn-go/dx/errors"
)

func (d *mySqlDialect) ParseError1062(err *mysql.MySQLError) *errors.DbErr {

	if strings.Contains(err.Message, "Duplicate entry ") {
		re := regexp.MustCompile(`for key '([^']+)'`)
		match := re.FindStringSubmatch(err.Message)
		if len(match) > 1 {
			//"for key 'users.UQ_users__username'"
			constraintName := match[1]
			if strings.Contains(constraintName, ".") {
				//"users.UQ_users__username"
				constraintName = strings.Split(constraintName, ".")[1]
			}
			ret := &errors.DbErr{
				Err:            err,
				ErrorType:      errors.ERR_DUPLICATE,
				ErrorMessage:   "duplicate",
				ConstraintName: constraintName,
			}
			ret.Reload()
			return ret
		} else {
			ret := &errors.DbErr{
				Err:          err,
				ErrorType:    errors.ERR_DUPLICATE,
				ErrorMessage: "duplicate",
			}
			ret.Reload()
			return ret
		}

	}
	return nil

}
