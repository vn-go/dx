package mysql

import "github.com/vn-go/dx/internal"

func (d *MysqlDialect) MakeSqlInsertBatch(tableName string, columns []internal.ColumnDef, data interface{}) (string, []interface{}) {
	panic("not implemented")
}
