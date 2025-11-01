package mssql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

// mssqlDialect.GetSelectStatement.go
func (d *mssqlDialect) GetSelectStatement(stmt types.SelectStatement) string {
	sql := ""
	if stmt.Limit != nil && stmt.Offset == nil {
		sql = "SELECT TOP (" + stmt.Limit.Content + ") " + stmt.Selector + " FROM " + stmt.Source
	} else {
		sql = "SELECT " + stmt.Selector + " FROM " + stmt.Source
	}

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

	// OFFSET + LIMIT (SQL Server yêu cầu ORDER BY nếu có OFFSET/FETCH)
	if stmt.Limit != nil && stmt.Offset != nil {
		if stmt.Sort == "" {
			// fallback để tránh lỗi "OFFSET requires ORDER BY clause"
			sql += " ORDER BY (SELECT NULL)"
		}

		sql += fmt.Sprintf(" OFFSET %s ROWS FETCH NEXT %s ROWS ONLY", stmt.Offset.Content, stmt.Limit.Content)

	}

	return sql
}
