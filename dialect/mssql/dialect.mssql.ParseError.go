package mssql

import (
	mssql "github.com/microsoft/go-mssqldb"
	"github.com/vn-go/dx/migate/loader/types"
)

func (d *mssqlDialect) ParseError(dbSchame *types.DbSchema, err error) error {
	//go-mssqldb.Error
	if mssqlErr, ok := err.(mssql.Error); ok {
		return d.Error2627(mssqlErr)
	}

	return err
}
