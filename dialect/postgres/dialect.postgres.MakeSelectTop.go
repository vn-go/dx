package postgres

import (
	"fmt"
)

func (d *postgresDialect) MakeSelectTop(query string, limit int) string {

	return query + " LIMIT " + fmt.Sprintf("%d", limit)
}
