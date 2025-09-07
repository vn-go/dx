package mssql

import (
	"fmt"
	"strings"
)

func (d *mssqlDialect) LimitAndOffset(sql string, limit, offset *uint64, orderBy string) string {

	if limit != nil && offset == nil {
		sql = strings.TrimPrefix(sql, "SELECT ")
		sql = "SELECT TOP " + fmt.Sprintf("%d", *limit) + " " + sql
		return sql
	}
	if limit != nil && offset == nil {
		return sql + fmt.Sprintf(" OFFSET %d", *limit)
	}
	if limit != nil && offset != nil {
		return sql + fmt.Sprintf(" OFFSET %d LIMIT %d", *offset, *limit)
	}
	return sql
}
