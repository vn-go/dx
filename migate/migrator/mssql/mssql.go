package mssql

import (
	"fmt"
	"reflect"
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

	db *db.DB
}

func NewMigrator(db *db.DB) migartorType.IMigrator {

	return &migratorMssql{
		db:     db,
		loader: loaderMysql.NewMysqlMigratorLoader(db),
	}
}
func (m *migratorMssql) Quote(names ...string) string {
	return "[" + strings.Join(names, "].[") + "]"
}

func (m *migratorMssql) GetSqlMigrateDelete(entityType reflect.Type) ([]string, error) {
	scripts := []string{}
	scriptTable, err := m.GetSqlCreateTable(entityType)
	if err != nil {
		return nil, err
	}
	if scriptTable == "" {
		scriptAddColumn, err := m.GetSqlAddColumn(entityType)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, scriptTable, scriptAddColumn)
	}

	scriptAddUniqueIndex, err := m.GetSqlAddUniqueIndex(entityType)
	if err != nil {
		return nil, err
	}
	scripts = append(scripts, scriptTable, scriptAddUniqueIndex)
	return scripts, nil

}
func (m *migratorMssql) DoMigrateDelete(entityType reflect.Type) error {
	scripts, err := m.GetSqlMigrateDelete(entityType)
	if err != nil {
		return err
	}
	for _, script := range scripts {
		_, err := m.db.Exec(script)
		if err != nil {
			return err
		}
	}
	return nil

}

type mssqlInitDoMigrates struct {
	once sync.Once
	err  error
}

var cacheDoMigrates sync.Map

func (m *migratorMssql) DoMigrates() error {

	key := fmt.Sprintf("%s_%s", m.db.DbName, m.db.DriverName)
	actual, _ := cacheDoMigrates.LoadOrStore(key, &mssqlInitDoMigrates{})

	mi := actual.(*mssqlInitDoMigrates)

	mi.once.Do(func() {

		scripts, err := m.GetFullScript()
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err := m.db.Exec(script)
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
