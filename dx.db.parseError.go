package dx

import (
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/migate/migrator"
)

func (db *DB) parseError(err error) error {
	dialect := factory.DialectFactory.Create(db.DriverName)

	imgrate, errGetMigator := migrator.GetMigator(db.DB)
	if errGetMigator != nil {
		return err
	}
	schema, errLoadFullSchema := imgrate.GetLoader().LoadFullSchema(db.DB)
	if errLoadFullSchema != nil {
		return err
	}

	return dialect.ParseError(schema, err)
}
