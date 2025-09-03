package mssql

import (
	"strings"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/common"

	mssql "github.com/microsoft/go-mssqldb"
)

func (d *MssqlDialect) Error2627(err mssql.Error) *errors.DbError {

	if strings.Contains(err.Message, "'") {
		constraint := strings.Split(err.Message, "'")[1]
		constraint = strings.Split(constraint, "'")[0]

		result := common.FindUKConstraint(constraint)
		if result != nil {
			cols := []string{}
			fields := []string{}
			for _, col := range result.Columns {
				cols = append(cols, col.Name)
				fields = append(fields, col.Field.Name)
			}
			ret := &errors.DbError{
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
