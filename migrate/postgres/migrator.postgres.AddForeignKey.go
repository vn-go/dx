package postgres

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/migrate/common"
)

func (m *MigratorPostgres) GetSqlAddForeignKey() ([]string, error) {
	ret := []string{}

	schema, err := m.Loader.LoadFullSchema(m.Db)
	if err != nil {
		return nil, err
	}

	for fk, info := range common.ForeignKeyRegistry.FKMap {
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
			schema.ForeignKeys[fk] = internal.DbForeignKeyInfo{
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
