package mssql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migate/loader/types"
	miragteTypes "github.com/vn-go/dx/migate/migrator/types"
)

func fkIfNotExists(constraintName, tableName, sqlFk string) string {
	ret := fmt.Sprintf(`
				IF NOT EXISTS (
					SELECT 1
					FROM sys.foreign_keys
					WHERE name = '%s'
					AND parent_object_id = OBJECT_ID('%s')
				)
				BEGIN
					%s;
				END
				`, constraintName, tableName, sqlFk)
	return ret
}
func (m *migratorMssql) GetSqlAddForeignKey(db *db.DB) ([]string, error) {
	ret := []string{}
	schema, err := m.loader.LoadFullSchema(db)
	if err != nil {
		return nil, err
	}

	for fk, info := range miragteTypes.ForeignKeyRegistry.FKMap {
		if _, ok := schema.ForeignKeys[strings.ToLower(fk)]; !ok {

			formCols := "[" + strings.Join(info.FromCols, "],[") + "]"
			toCols := "[" + strings.Join(info.ToCols, "],[") + "]"
			script := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)", m.Quote(info.FromTable), m.Quote(fk), formCols, m.Quote(info.ToTable), toCols)
			schema.ForeignKeys[strings.ToLower(fk)] = types.DbForeignKeyInfo{
				ConstraintName: fk,
				Table:          info.ToTable,
				Columns:        info.FromCols,
				RefTable:       info.ToTable,
				RefColumns:     info.ToCols,
			}
			if info.Cascade.OnDelete {
				script += " ON DELETE CASCADE"
			}
			if info.Cascade.OnUpdate {
				script += " ON UPDATE CASCADE"
			}
			scriptIfNotExit := fkIfNotExists(fk, info.FromTable, script)
			ret = append(ret, scriptIfNotExit)
		}
	}
	return ret, nil
}
