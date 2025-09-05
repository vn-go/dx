package mysql

import "github.com/vn-go/dx/entity"

func (d *mySqlDialect) MakeSqlInsertBatch(tableName string, columns []entity.ColumnDef, data interface{}) (string, []interface{}) {
	panic("not implemented")
}
