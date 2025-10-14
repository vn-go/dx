package mssql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	dxErrors "github.com/vn-go/dx/errors"
	loaderMssql "github.com/vn-go/dx/migrate/loader/mssql"
	"github.com/vn-go/dx/migrate/loader/types"
	migartorType "github.com/vn-go/dx/migrate/migrator/types"
)

type migratorMssql struct {
	loader             types.IMigratorLoader
	cacheGetFullScript sync.Map
}

func NewMigrator() migartorType.IMigrator {

	return &migratorMssql{

		loader: &loaderMssql.MigratorLoaderMssql{},
	}
}
func (m *migratorMssql) Quote(names ...string) string {
	return "[" + strings.Join(names, "].[") + "]"
}
func (m *migratorMssql) GetDefaultSchema() string {
	return "dbo"
}

type mssqlInitDoMigrates struct {
	once sync.Once
	err  error
}

var cacheDoMigrates sync.Map

func (m *migratorMssql) DoMigrates(db *db.DB, schema string) error {

	key := fmt.Sprintf("%s_%s", db.DbName, db.DriverName)
	actual, _ := cacheDoMigrates.LoadOrStore(key, &mssqlInitDoMigrates{})

	mi := actual.(*mssqlInitDoMigrates)

	mi.once.Do(func() {

		scripts, err := m.GetFullScript(db, schema)
		if err != nil {
			mi.err = err
			return
		}
		for _, script := range scripts {
			_, err := db.Exec(script)
			if err != nil {
				mi.err = dxErrors.NewMigrationError(script, err)
				break
			}
		}
		// for _, entity := range ModelRegistry.GetAllModels() {
		// 	err = m.DoMigrate(entity.entity.entityType)

		// }
	})
	return mi.err
}

func (m *migratorMssql) GetLoader() types.IMigratorLoader {
	return m.loader
}
