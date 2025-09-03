package mysql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/migrate/common"
)

func (m *MigratorMySql) GetSqlAddForeignKey() ([]string, error) {
	ret := []string{}

	schema, err := m.Loader.LoadFullSchema(m.Db)
	if err != nil {
		return nil, err
	}

	for fk, info := range common.ForeignKeyRegistry.FKMap {
		if _, ok := schema.ForeignKeys[fk]; !ok {

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
			schema.ForeignKeys[fk] = common.DbForeignKeyInfo{
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
