package migrate

import (
	"fmt"

	common "github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/migrate/mssql"
	"github.com/vn-go/dx/migrate/mysql"
	"github.com/vn-go/dx/migrate/postgres"
	"github.com/vn-go/dx/tenantDB"
)

func MigratorLoader(db *tenantDB.TenantDB) (common.IMigratorLoader, error) {
	err := db.Detect()
	if err != nil {
		return nil, err
	}
	switch db.GetDbType() {
	case tenantDB.DB_DRIVER_MSSQL:
		return &mssql.MigratorLoaderMssql{}, nil
	case tenantDB.DB_DRIVER_Postgres:
		return &postgres.MigratorLoaderPostgres{}, nil
	case tenantDB.DB_DRIVER_MySQL:
		return &mysql.MigratorLoaderMysql{}, nil

	default:
		panic(fmt.Errorf("unsupported database type: %s", string(db.GetDbType())))
	}
}
