package postgres

import (
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migate/loader/types"

	"github.com/lib/pq"
)

func (d *postgresDialect) ParseError42601(dbSchame *types.DbSchema, err *pq.Error) error {
	return &errors.DbErr{
		// Table:        err.Table,
		ErrorType: errors.ERR_SYNTAX,
		// DbCols:       []string{err.Column},
		Err:          err,
		ErrorMessage: err.Message,
	}

}
