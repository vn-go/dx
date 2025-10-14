package postgres

import (
	"fmt"

	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/loader/types"

	"github.com/lib/pq"
)

func (d *postgresDialect) ParseError23502(dbSchame *types.DbSchema, err *pq.Error) error {
	return &errors.DbErr{
		Table:        err.Table,
		ErrorType:    errors.ERR_REQUIRED,
		DbCols:       []string{err.Column},
		Err:          err,
		ErrorMessage: fmt.Sprintf("%s.%s can not be null", err.Table, err.Column),
	}
	// ukContraint := err.Constraint
	// if colsInfo, ok := dbSchame.UniqueKeys[ukContraint]; ok {
	// 	dbCols := []string{}
	// 	for _, col := range colsInfo.Columns {
	// 		dbCols = append(dbCols, col.Name)
	// 	}
	// 	return &errors.DbErr{
	// 		Err:            err,
	// 		ErrorType:      errors.ERR_DUPLICATE,
	// 		DbCols:         dbCols,
	// 		Table:          colsInfo.TableName,
	// 		Fields:         dbCols,
	// 		ErrorMessage:   "duplicate",
	// 		ConstraintName: ukContraint,
	// 	}
	// }

	//return err
}
