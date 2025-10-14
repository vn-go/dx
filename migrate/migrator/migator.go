package migrator

import (
	"fmt"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/migrate/migrator/mssql"
	"github.com/vn-go/dx/migrate/migrator/mysql"
	"github.com/vn-go/dx/migrate/migrator/oracle"
	"github.com/vn-go/dx/migrate/migrator/postgres"
	"github.com/vn-go/dx/migrate/migrator/types"
)

func GetMigator(db *db.DB) (types.IMigrator, error) {
	ret, err := internal.OnceCall("Migator/GetMigator/"+db.DbName+"/"+db.DriverName, func() (types.IMigrator, error) {
		if db.Info.DriverName == "mysql" {
			return mysql.NewMigrator(), nil
		}
		if db.Info.DriverName == "sqlserver" {
			return mssql.NewMigrator(), nil
		}
		if db.Info.DriverName == "postgres" {
			return postgres.NewMigrator(), nil
		}
		if db.Info.DriverName == "oracle" {
			return oracle.NewMigrator(), nil
		}
		panic(fmt.Sprintf("%s is not supported in func db.DB.GetMigator. source file \"%s\" ", db.Info.DriverName, `migrate\migrator\migator.go`))
		//return nil, errors.NewUnsupportedError(fmt.Sprintf("%s is not supported", db.Info.DriverName))
	})
	return ret, err

}
