package mysql

import "fmt"

func (d *MysqlDialect) MakeSelectTop(query string, limit int) string {
	return query + " LIMIT " + fmt.Sprintf("%d", limit)
}
