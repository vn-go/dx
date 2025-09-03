package postgres

import (
	"fmt"
)

func (d *PostgresDialect) MakeSelectTop(query string, limit int) string {

	return query + " LIMIT " + fmt.Sprintf("%d", limit)
}
