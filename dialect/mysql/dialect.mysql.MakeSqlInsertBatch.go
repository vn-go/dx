package mysql

import (
	"fmt"

	"github.com/vn-go/dx/entity"
)

func (d *mySqlDialect) MakeSqlInsertBatch(tableName string, columns []entity.ColumnDef, data interface{}) (string, []interface{}) {
	panic(fmt.Sprintf("not implemented MakeSqlInsertBatch, %s", `dialect\mysql\dialect.mysql.MakeSqlInsertBatch.go`))
}
