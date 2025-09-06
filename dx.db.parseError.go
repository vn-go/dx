package dx

import (
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/migate/migrator"
)

func (db *DB) parseError(err error) error {
	dialect := factory.DialectFactory.Create(db.DriverName)

	if err != nil {
		return err
	}
	imgrate, err := migrator.GetMigator(db.DB)
	if err != nil {
		return err
	}
	schema, err := imgrate.GetLoader().LoadFullSchema(db.DB)
	if err != nil {
		return err
	}

	return dialect.ParseError(schema, err)
}
