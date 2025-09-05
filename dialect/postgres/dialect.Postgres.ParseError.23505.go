package postgres

import (
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migate/loader/types"

	"github.com/lib/pq"
)

func (d *postgresDialect) ParseError23505(dbSchame *types.DbSchema, err *pq.Error) error {
	ukContraint := err.Constraint
	if colsInfo, ok := dbSchame.UniqueKeys[ukContraint]; ok {
		dbCols := []string{}
		for _, col := range colsInfo.Columns {
			dbCols = append(dbCols, col.Name)
		}
		return &errors.DbErr{
			Err:            err,
			ErrorType:      errors.ERR_DUPLICATE,
			DbCols:         dbCols,
			Table:          colsInfo.TableName,
			Fields:         dbCols,
			ErrorMessage:   "duplicate",
			ConstraintName: ukContraint,
		}
	}

	return err
}
