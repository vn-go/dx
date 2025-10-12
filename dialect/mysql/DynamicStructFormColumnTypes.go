package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
)

func (d *mySqlDialect) DynamicStructFormColumnTypes(sql string, colTypes []*sql.ColumnType) reflect.Type {
	panic(fmt.Sprintf("Not impeleted mySqlDialect.DynamicStructFormSqlColumns,%s", `dialect\mysql\DynamicStructFormColumnTypes.go`))

}
