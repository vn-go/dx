package postgres

import (
	"fmt"
)

func (d *postgresDialect) LimitAndOffset(sql string, limit, offset *uint64, orderBy string) string {
	if orderBy != "" {
		sql += " ORDER BY " + orderBy
	}
	if limit != nil && offset == nil {
		return sql + fmt.Sprintf(" LIMIT  %d", *limit)
	}
	if limit == nil && offset != nil {
		return sql + fmt.Sprintf(" OFFSET %d", *offset)
	}
	if limit != nil && offset != nil {
		return sql + fmt.Sprintf(" OFFSET %d LIMIT %d", *offset, *limit)
	}
	return sql
}
