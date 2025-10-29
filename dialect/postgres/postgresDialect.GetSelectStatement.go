package postgres

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

// postgresDialect.GetSelectStatement.go
func (d *postgresDialect) GetSelectStatement(stmt types.SelectStatement) string {
	sql := "SELECT " + stmt.Selector + " FROM " + stmt.Source

	// WHERE
	if stmt.Filter != "" {
		sql += " WHERE " + stmt.Filter
	}

	// GROUP BY
	if stmt.GroupBy != "" {
		sql += " GROUP BY " + stmt.GroupBy
	}

	// HAVING
	if stmt.Having != "" {
		sql += " HAVING " + stmt.Having
	}

	// ORDER BY
	if stmt.Sort != "" {
		sql += " ORDER BY " + stmt.Sort
	}

	// LIMIT & OFFSET
	if stmt.Limit != nil {
		sql += fmt.Sprintf(" LIMIT %s", stmt.Limit.Content)
		if stmt.Offset != nil {
			sql += fmt.Sprintf(" OFFSET %s", stmt.Offset.Content)
		}
	}

	return sql
}
