package mssql

import (
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/internal"
)

func (d *MssqlDialect) ParseError(dbSchame *internal.DbSchema, err error) error {
	//go-mssqldb.Error
	if mssqlErr, ok := err.(mssql.Error); ok {
		return d.Error2627(mssqlErr)
	}

	return err
}
