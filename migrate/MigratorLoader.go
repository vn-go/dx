package migrate

import (
	"database/sql"
	"fmt"

	"github.com/vn-go/dx/internal"
	common "github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/migrate/mssql"
	"github.com/vn-go/dx/migrate/mysql"
	"github.com/vn-go/dx/migrate/postgres"
)

func MigratorLoader(db *sql.DB) (common.IMigratorLoader, error) {
	err := db.Detect()
	if err != nil {
		return nil, err
	}
	switch db.GetDbType() {
	case internal.DB_DRIVER_MSSQL:
		return &mssql.MigratorLoaderMssql{}, nil
	case internal.DB_DRIVER_Postgres:
		return &postgres.MigratorLoaderPostgres{}, nil
	case internal.DB_DRIVER_MySQL:
		return &mysql.MigratorLoaderMysql{}, nil

	default:
		panic(fmt.Errorf("unsupported database type: %s", string(db.GetDbType())))
	}
}
