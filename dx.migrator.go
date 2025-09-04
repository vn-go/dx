package dx

import (
	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migate/migrator"
)

type migratorType struct {
}

func (m *migratorType) GetSql(db *db.DB) ([]string, error) {
	migate, err := migrator.GetMigator(db)
	if err != nil {
		return nil, err
	}
	return migate.GetFullScript()
}

var Migrator = &migratorType{}
