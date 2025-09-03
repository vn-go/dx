package mysql

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

type MigratorMySql struct {
	Loader             common.IMigratorLoader
	cacheGetFullScript sync.Map

	Db *tenantDB.TenantDB
}

func (m *MigratorMySql) GetLoader() common.IMigratorLoader {
	return m.Loader
}
func (m *MigratorMySql) Quote(names ...string) string {
	return "`" + strings.Join(names, "`.`") + "`"
}

func (m *MigratorMySql) GetSqlMigrate(entityType reflect.Type) ([]string, error) {
	panic("implement me")
}

func (m *MigratorMySql) DoMigrate(entityType reflect.Type) error {
	panic("implement me")
}

type mysqlInitDoMigrates struct {
	once sync.Once
	err  error
}

var cacheMysqlDoMigrates sync.Map

func (m *MigratorMySql) DoMigrates() error {

	key := fmt.Sprintf("%s_%s", m.Db.GetDBName(), m.Db.GetDbType())
	actual, _ := cacheMysqlDoMigrates.LoadOrStore(key, &mysqlInitDoMigrates{})

	mi := actual.(*mysqlInitDoMigrates)

	mi.once.Do(func() {

		scripts, err := m.GetFullScript()
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err := m.Db.Exec(script)
			if err != nil {
				mi.err = errors.CreateError(script, err)
				break
			}
		}

	})
	return mi.err
}
