package mysql

import "fmt"

// for sql server
//
// SELECT * FROM [users] ORDER BY (SELECT NULL) OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY;
func (d *mySqlDialect) LimitAndOffset(sql string, limit, offset *uint64, orderBy string) string {
	if orderBy != "" {
		sql += " ORDER BY " + orderBy
	}
	if limit != nil && offset == nil {
		return sql + fmt.Sprintf(" LIMIT  %d", *limit)
	}
	if limit == nil && offset != nil {
		return sql + fmt.Sprintf(" OFFSET %d", *limit)
	}
	if limit != nil && offset != nil {
		return sql + fmt.Sprintf(" OFFSET %d LIMIT %d", *offset, *limit)
	}

	return sql
}
