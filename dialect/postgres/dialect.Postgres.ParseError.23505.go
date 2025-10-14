package postgres

import (
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/loader/types"

	"github.com/lib/pq"
)

/*
For PostgreSQL error 23505, it is necessary to check the UniqueKeys and PrimaryKeys in the database schema for detailed inspection
*/
func (d *postgresDialect) ParseError23505(dbSchame *types.DbSchema, err *pq.Error) error {
	//"23505"
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
	if colsInfo, ok := dbSchame.PrimaryKeys[ukContraint]; ok {
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
