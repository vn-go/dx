package mssql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	loaderMysql "github.com/vn-go/dx/migate/loader/mysql"
	"github.com/vn-go/dx/migate/loader/types"
	migartorType "github.com/vn-go/dx/migate/migrator/types"
)

type migratorMssql struct {
	loader             types.IMigratorLoader
	cacheGetFullScript sync.Map
}

func NewMigrator() migartorType.IMigrator {

	return &migratorMssql{

		loader: loaderMysql.NewMysqlMigratorLoader(),
	}
}
func (m *migratorMssql) Quote(names ...string) string {
	return "[" + strings.Join(names, "].[") + "]"
}

type mssqlInitDoMigrates struct {
	once sync.Once
	err  error
}

var cacheDoMigrates sync.Map

func (m *migratorMssql) DoMigrates(db *db.DB) error {

	key := fmt.Sprintf("%s_%s", db.DbName, db.DriverName)
	actual, _ := cacheDoMigrates.LoadOrStore(key, &mssqlInitDoMigrates{})

	mi := actual.(*mssqlInitDoMigrates)

	mi.once.Do(func() {

		scripts, err := m.GetFullScript(db)
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err := db.Exec(script)
			if err != nil {
				mi.err = err
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
