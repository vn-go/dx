package postgres

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/vn-go/dx/internal"
)

type initDynamicStructFormColumnTypes struct {
	val  reflect.Type
	once sync.Once
}

var initDynamicStructFormColumnTypesCache sync.Map

func (d *postgresDialect) DynamicStructFormColumnTypes(sql string, colTypes []*sql.ColumnType) reflect.Type {
	a, _ := initDynamicStructFormColumnTypesCache.LoadOrStore(sql, &initDynamicStructFormColumnTypes{})
	i := a.(*initDynamicStructFormColumnTypes)
	i.once.Do(func() {
		i.val = d.dynamicStructFormColumnTypes(colTypes)
	})
	//panic(fmt.Sprintf("Not impeleted postgresDialect.DynamicStructFormSqlColumns,%s", `dialect\postgres\DynamicStructFormColumnTypes.go`))
	return i.val
}
func (d *postgresDialect) dynamicStructFormColumnTypes(colTypes []*sql.ColumnType) reflect.Type {
	fields := make([]reflect.StructField, 0, len(colTypes))

	for _, col := range colTypes {
		dbType := strings.ToUpper(col.DatabaseTypeName())
		goType := reflect.TypeOf(new(interface{})).Elem() // fallback

		switch dbType {
		case "INT", "INTEGER", "SMALLINT", "SERIAL", "BIGSERIAL":
			goType = reflect.TypeFor[*int64]()
		case "BIGINT":
			goType = reflect.TypeFor[*int64]()
		case "REAL", "FLOAT4":
			goType = reflect.TypeOf(float32(0))
		case "DOUBLE", "FLOAT8", "NUMERIC", "DECIMAL":
			goType = reflect.TypeFor[*float64]()
		case "BOOLEAN", "BOOL":
			goType = reflect.TypeFor[*bool]()
		case "CHAR", "VARCHAR", "TEXT", "CITEXT", "UUID":
			goType = reflect.TypeOf("")
		case "DATE", "TIMESTAMP", "TIMESTAMPTZ":
			goType = reflect.TypeFor[*time.Time]() //reflect.TypeOf(time.Time{})
		case "BYTEA":
			goType = reflect.TypeOf([]byte(nil))
		default:
			goType = reflect.TypeOf(new(interface{})).Elem()
		}

		field := reflect.StructField{
			Name: strings.Title(col.Name()), // viết hoa để export
			Type: goType,
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, internal.Helper.ToLowerCamel(col.Name()))),
		}
		fields = append(fields, field)
	}

	return reflect.StructOf(fields)
}
