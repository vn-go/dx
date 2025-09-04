package mysql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	loaderMssql "github.com/vn-go/dx/migate/loader/mssql"
	"github.com/vn-go/dx/migate/loader/types"
	migartorType "github.com/vn-go/dx/migate/migrator/types"
)

type MigratorMySql struct {
	loader             types.IMigratorLoader
	cacheGetFullScript sync.Map

	db *db.DB
}

func NewMigrator(db *db.DB) migartorType.IMigrator {

	return &MigratorMySql{
		db:     db,
		loader: loaderMssql.NewMssqlSchemaLoader(db),
	}
}
func (m *MigratorMySql) GetLoader() types.IMigratorLoader {
	return m.loader
}
func (m *MigratorMySql) Quote(names ...string) string {
	return "`" + strings.Join(names, "`.`") + "`"
}

type mssqlInitDoMigrates struct {
	once sync.Once
	err  error
}

var cacheMigratorMySqlDoMigrates sync.Map

func (m *MigratorMySql) DoMigrates() error {

	key := fmt.Sprintf("%s_%s", m.db.DbName, m.db.DriverName)
	actual, _ := cacheMigratorMySqlDoMigrates.LoadOrStore(key, &mssqlInitDoMigrates{})

	mi := actual.(*mssqlInitDoMigrates)

	mi.once.Do(func() {

		scripts, err := m.GetFullScript()
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err := m.db.Exec(script)
			if err != nil {
				mi.err = errors.NewMigrationError(script, err)
				break
			}
		}

	})
	return mi.err
}
