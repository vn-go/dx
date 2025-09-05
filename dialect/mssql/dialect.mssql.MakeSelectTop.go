package mssql

import (
	"fmt"
	"strings"
)

func (d *mssqlDialect) MakeSelectTop(query string, limit int) string {
	query = strings.TrimPrefix(query, "SELECT ")
	query = "SELECT TOP " + fmt.Sprintf("%d", limit) + " " + query
	return query
}
