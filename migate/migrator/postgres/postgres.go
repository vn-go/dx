package postgres

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	loaderPostgres "github.com/vn-go/dx/migate/loader/postgres"
	"github.com/vn-go/dx/migate/loader/types"
	migartorType "github.com/vn-go/dx/migate/migrator/types"
)

type MigratorPostgres struct {
	loader             types.IMigratorLoader
	cacheGetFullScript sync.Map
}

func NewMigrator() migartorType.IMigrator {

	return &MigratorPostgres{

		loader: loaderPostgres.NewPosgresSchemaLoader(),
	}
}
func (m *MigratorPostgres) GetLoader() types.IMigratorLoader {
	return m.loader
}
func (m *MigratorPostgres) Quote(names ...string) string {
	return "\"" + strings.Join(names, "\".\"") + "\""
}

type postgresInitDoMigrates struct {
	once sync.Once
}

var cacheMigratorPostgresMigratorPostgres sync.Map

func (m *MigratorPostgres) DoMigrates(db *db.DB) error {
	key := fmt.Sprintf("%s_%s", db.DbName, db.DriverName)
	actual, _ := cacheMigratorPostgresMigratorPostgres.LoadOrStore(key, &postgresInitDoMigrates{})

	mi := actual.(*postgresInitDoMigrates)
	var err error
	mi.once.Do(func() {
		var scripts []string
		scripts, err = m.GetFullScript(db)
		if err != nil {
			return
		}
		for _, script := range scripts {
			_, err = db.Exec(script)
			if err != nil {
				err = fmt.Errorf("sql-error:\n%s\n%s", script, err.Error())
				break
			}
		}

	})
	return err
}
