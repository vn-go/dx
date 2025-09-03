package postgres

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

type MigratorPostgres struct {
	Loader             common.IMigratorLoader
	cacheGetFullScript sync.Map

	Db *tenantDB.TenantDB
}

func (m *MigratorPostgres) GetLoader() common.IMigratorLoader {
	return m.Loader
}
func (m *MigratorPostgres) Quote(names ...string) string {
	return "\"" + strings.Join(names, "\".\"") + "\""
}

func (m *MigratorPostgres) GetSqlMigrate(entityType reflect.Type) ([]string, error) {
	panic("not implemented")
}

func (m *MigratorPostgres) DoMigrate(entityType reflect.Type) error {
	panic("not implemented")
}

type postgresInitDoMigrates struct {
	once sync.Once
}

var cacheDoMigrates sync.Map

func (m *MigratorPostgres) DoMigrates() error {
	key := fmt.Sprintf("%s_%s", m.Db.GetDBName(), m.Db.GetDbType())
	actual, _ := cacheDoMigrates.LoadOrStore(key, &postgresInitDoMigrates{})

	mi := actual.(*postgresInitDoMigrates)
	var err error
	mi.once.Do(func() {
		var scripts []string
		scripts, err = m.GetFullScript()
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err = m.Db.Exec(script)
			if err != nil {
				err = fmt.Errorf("sql-error:\n%s\n%s", script, err.Error())
				break
			}
		}

	})
	return err
}
