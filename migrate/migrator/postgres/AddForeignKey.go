package postgres

import (
	"fmt"

	"strings"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
	miragteTypes "github.com/vn-go/dx/migrate/migrator/types"
)

func addFkIfNotExist(constraintName, tableName, sqlAddFK string, schema string) string {
	ret := fmt.Sprintf(`
		DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1
					FROM pg_constraint
					WHERE conname = '%s'
					AND conrelid = '%s'::regclass
				) THEN
					%s;
			END IF;
		END$$;
		`, constraintName, tableName, sqlAddFK)
	return ret

}
func (m *MigratorPostgres) GetSqlAddForeignKey(db *db.DB, shema string) ([]string, error) {
	ret := []string{}

	schema, err := m.loader.LoadFullSchema(db, shema)
	if err != nil {
		return nil, err
	}

	for fk, info := range miragteTypes.ForeignKeyRegistry.FKMap {
		if _, ok := schema.ForeignKeys[fk]; !ok {
			// Quote cột
			fromCols := []string{}
			for _, col := range info.FromCols {
				fromCols = append(fromCols, m.Quote(col))
			}
			toCols := []string{}
			for _, col := range info.ToCols {
				toCols = append(toCols, m.Quote(col))
			}

			scriptAddFK := fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
				m.Quote(info.FromTable),
				m.Quote(fk),
				strings.Join(fromCols, ", "),
				m.Quote(info.ToTable),
				strings.Join(toCols, ", "),
			)

			if info.Cascade.OnDelete {
				scriptAddFK += " ON DELETE CASCADE"
			}
			if info.Cascade.OnUpdate {
				scriptAddFK += " ON UPDATE CASCADE"
			}
			script := addFkIfNotExist(fk, info.FromTable, scriptAddFK, shema)
			ret = append(ret, script)

			// Cập nhật lại schema cache
			schema.ForeignKeys[fk] = types.DbForeignKeyInfo{
				ConstraintName: fk,
				Table:          info.FromTable,
				Columns:        info.FromCols,
				RefTable:       info.ToTable,
				RefColumns:     info.ToCols,
			}
		}
	}

	return ret, nil
}
