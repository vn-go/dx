package mysql

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/errors"
	loaderMysql "github.com/vn-go/dx/migate/loader/mysql"
	"github.com/vn-go/dx/migate/loader/types"
	migartorType "github.com/vn-go/dx/migate/migrator/types"
)

type MigratorMySql struct {
	loader             types.IMigratorLoader
	cacheGetFullScript sync.Map
}

func NewMigrator() migartorType.IMigrator {

	return &MigratorMySql{

		loader: loaderMysql.NewMysqlMigratorLoader(),
	}
}
func (m *MigratorMySql) GetLoader() types.IMigratorLoader {
	return m.loader
}
func (m *MigratorMySql) Quote(names ...string) string {
	return "`" + strings.Join(names, "`.`") + "`"
}

type mysqlInitDoMigrates struct {
	once sync.Once
	err  error
}

var cacheMigratorMySqlDoMigrates sync.Map

func (m *MigratorMySql) DoMigrates(db *db.DB) error {

	key := fmt.Sprintf("%s_%s", db.DbName, db.DriverName)
	actual, _ := cacheMigratorMySqlDoMigrates.LoadOrStore(key, &mysqlInitDoMigrates{})

	mi := actual.(*mysqlInitDoMigrates)

	mi.once.Do(func() {

		scripts, err := m.GetFullScript(db)
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err := db.Exec(script)
			if err != nil {
				mi.err = errors.NewMigrationError(script, err)
				break
			}
		}

	})
	return mi.err
}
