package mssql

import (
	"strings"

	"github.com/vn-go/dx/errors"

	_ "github.com/microsoft/go-mssqldb"

	mssql "github.com/microsoft/go-mssqldb"
)

func (d *mssqlDialect) Error2627(err mssql.Error) *errors.DbErr {

	if strings.Contains(err.Message, "'") {
		constraint := strings.Split(err.Message, "'")[1]
		constraint = strings.Split(constraint, "'")[0]

		result := FindUKConstraint(constraint)
		if result != nil {
			cols := []string{}
			fields := []string{}
			for _, col := range result.Columns {
				cols = append(cols, col.Name)
				fields = append(fields, col.Field.Name)
			}
			ret := &errors.DbErr{
				Err:          err,
				ErrorType:    errors.ERR_DUPLICATE,
				ErrorMessage: err.Message,
				DbCols:       cols,
				Fields:       fields,
				Table:        result.TableName,
			}
			ret.Reload()
			return ret
		}

	}
	// errorMsg := err.Message
	panic("not implemented")
}
