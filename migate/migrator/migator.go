package migrator

import (
	"fmt"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migate/migrator/mysql"
	"github.com/vn-go/dx/migate/migrator/postgres"
	"github.com/vn-go/dx/migate/migrator/types"
)

func GetMigator(db *db.DB) (types.IMigrator, error) {
	if db.Info.DriverName == "mysql" {
		return mysql.NewMigrator(db), nil
	}
	if db.Info.DriverName == "sqlserver" {
		return mysql.NewMigrator(db), nil
	}
	if db.Info.DriverName == "postgres" {
		return postgres.NewMigrator(db), nil
	}
	return nil, errors.NewUnsupportedError(fmt.Sprintf("%s is not supported", db.Info.DriverName))

}
