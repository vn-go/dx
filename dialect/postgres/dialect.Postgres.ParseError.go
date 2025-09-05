package postgres

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/vn-go/dx/migate/loader/types"
)

func (d *postgresDialect) ParseError(dbSchema *types.DbSchema, err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return d.ParseError23505(dbSchema, pgErr)

		}
		panic(fmt.Errorf(`not implemented,vdb\dialect.Postgres.go`))
	} else {
		return err
	}

}
