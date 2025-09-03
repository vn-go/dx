package migrate

import (
	"fmt"
	"sync"

	common "github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/migrate/mssql"
	"github.com/vn-go/dx/migrate/mysql"
	"github.com/vn-go/dx/migrate/postgres"
	"github.com/vn-go/dx/tenantDB"
)

var cacheNewMigrator sync.Map

type initNewMigrator struct {
	once sync.Once
	err  error
	val  common.IMigrator
}

func NewMigrator(db *tenantDB.TenantDB) (common.IMigrator, error) {
	err := db.Detect()
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s:%s", db.GetDBName(), db.GetDbType())
	actual, _ := cacheNewMigrator.LoadOrStore(key, &initNewMigrator{})
	mi := actual.(*initNewMigrator)
	mi.once.Do(func() {

		loader, err := MigratorLoader(db)
		if err != nil {
			mi.err = err
		}
		switch db.GetDbType() {
		case tenantDB.DB_DRIVER_MSSQL:

			mi.val = &mssql.MigratorMssql{
				Db:     db,
				Loader: loader,
			}
		case tenantDB.DB_DRIVER_Postgres:
			mi.val = &postgres.MigratorPostgres{
				Db:     db,
				Loader: loader,
			}
		case tenantDB.DB_DRIVER_MySQL:
			mi.val = &mysql.MigratorMySql{
				Db:     db,
				Loader: loader,
			}
		default:
			mi.err = fmt.Errorf("unsupported database type: %s", db.GetDbType())
		}
	})
	return mi.val, mi.err
}
