package mssql

import (
	"database/sql"
	"fmt"
	"reflect"
)

func (d *mssqlDialect) DynamicStructFormColumnTypes(sql string, colTypes []*sql.ColumnType) reflect.Type {
	panic(fmt.Sprintf("Not impeleted mssqlDialect.DynamicStructFormSqlColumns,%s", `dialect\mssql\DynamicStructFormColumnTypes.go`))
}
