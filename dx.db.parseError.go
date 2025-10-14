package dx

import (
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/migrate/migrator"
)

func (db *DB) parseError(err error) error {
	dialect := factory.DialectFactory.Create(db.DriverName)

	imgrate, errGetMigator := migrator.GetMigator(db.DB)
	if errGetMigator != nil {
		return err
	}
	schema, errLoadFullSchema := imgrate.GetLoader().LoadFullSchema(db.DB, imgrate.GetDefaultSchema())
	if errLoadFullSchema != nil {
		return err
	}

	return dialect.ParseError(schema, err)
}
