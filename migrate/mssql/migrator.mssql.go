package mssql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	common "github.com/vn-go/dx/migrate/common"
)

type MigratorMssql struct {
	Loader             common.IMigratorLoader
	cacheGetFullScript sync.Map

	Db *sql.DB
}

func (m *MigratorMssql) Quote(names ...string) string {
	return "[" + strings.Join(names, "].[") + "]"
}

func (m *MigratorMssql) GetSqlMigrate(entityType reflect.Type) ([]string, error) {
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
func (m *MigratorMssql) DoMigrate(entityType reflect.Type) error {
	scripts, err := m.GetSqlMigrate(entityType)
	if err != nil {
		return err
	}
	for _, script := range scripts {
		_, err := m.Db.Exec(script)
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

func (m *MigratorMssql) DoMigrates(dbName, dbType string) error {

	key := fmt.Sprintf("%s_%s", dbName, dbType)
	actual, _ := cacheDoMigrates.LoadOrStore(key, &mssqlInitDoMigrates{})

	mi := actual.(*mssqlInitDoMigrates)

	mi.once.Do(func() {

		scripts, err := m.GetFullScript(dbName, dbType)
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err := m.Db.Exec(script)
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

func (m *MigratorMssql) GetLoader() common.IMigratorLoader {
	return m.Loader
}
