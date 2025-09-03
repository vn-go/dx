package dx

import "github.com/vn-go/dx/migrate/common"

func (d *MysqlDialect) MakeSqlInsertBatch(tableName string, columns []common.ColumnDef, data interface{}) (string, []interface{}) {
	panic("not implemented")
}
