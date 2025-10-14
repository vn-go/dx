package postgres

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	dxErrors "github.com/vn-go/dx/errors"
	loaderPostgres "github.com/vn-go/dx/migrate/loader/postgres"
	"github.com/vn-go/dx/migrate/loader/types"
	migartorType "github.com/vn-go/dx/migrate/migrator/types"
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
	err  error
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
			// subScript := strings.Split(script, ";")
			_, err = db.Exec(script)
			if err != nil {
				err = dxErrors.NewMigrationError(script, err)

				return
			}

		}

	})
	return err
}
