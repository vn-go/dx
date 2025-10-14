package postgres

import (
	"errors"
	"fmt"

	"github.com/lib/pq"
	"github.com/vn-go/dx/migrate/loader/types"
)

func (d *postgresDialect) ParseError(dbSchema *types.DbSchema, err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return d.ParseError23505(dbSchema, pgErr)

		}
		if pgErr.Code == "23502" {
			return d.ParseError23502(dbSchema, pgErr)

		}
		if pgErr.Code == "42601" {
			return d.ParseError42601(dbSchema, pgErr)

		}
		if pgErr.Code == "42P18" {
			return d.ParseError42P18(dbSchema, pgErr)

		}
		if pgErr.Code == "22P02" {
			return pgErr
		}
		if pgErr.Code == "42P02" {
			return pgErr
		}
		if pgErr.Code == "42803" {
			return pgErr
		}
		panic(fmt.Errorf(`not implemented error code %s at %s`, pgErr.Code, `dialect\postgres\dialect.Postgres.ParseError.go`))
	} else {
		return err
	}

}
