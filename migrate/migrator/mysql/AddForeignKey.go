package mysql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/db"
	loadeTypes "github.com/vn-go/dx/migrate/loader/types"
	miragteTypes "github.com/vn-go/dx/migrate/migrator/types"
)

func (m *MigratorMySql) GetSqlAddForeignKey(db *db.DB) ([]string, error) {
	ret := []string{}

	schema, err := m.loader.LoadFullSchema(db)
	if err != nil {
		return nil, err
	}

	for fk, info := range miragteTypes.ForeignKeyRegistry.FKMap {
		if _, ok := schema.ForeignKeys[strings.ToLower(fk)]; !ok {

			// Quote column names bằng backtick cho MySQL
			fromCols := make([]string, len(info.FromCols))
			toCols := make([]string, len(info.ToCols))
			for i, c := range info.FromCols {
				fromCols[i] = m.Quote(c)
			}
			for i, c := range info.ToCols {
				toCols[i] = m.Quote(c)
			}

			script := fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
				m.Quote(info.FromTable),
				m.Quote(fk),
				strings.Join(fromCols, ", "),
				m.Quote(info.ToTable),
				strings.Join(toCols, ", "),
			)

			if info.Cascade.OnDelete {
				script += " ON DELETE CASCADE"
			}
			if info.Cascade.OnUpdate {
				script += " ON UPDATE CASCADE"
			}

			ret = append(ret, script)

			// Cập nhật lại schema cache
			schema.ForeignKeys[fk] = loadeTypes.DbForeignKeyInfo{
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
