package postgres

import (
	"github.com/lib/pq"
	"github.com/vn-go/dx/migate/loader/types"
)

func (d *postgresDialect) ParseError42P18(dbSchame *types.DbSchema, err *pq.Error) error {
	return err
}
