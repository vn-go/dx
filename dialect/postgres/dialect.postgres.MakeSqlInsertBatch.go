package postgres

import (
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/internal"
)

type makePostgresSqlInsertInit struct {
	once sync.Once
	val  string
}

// MakeSqlInsert(tableName string, columns []internal.ColumnDef, data interface{}) (string, []any)
func (d *PostgresDialect) MakeSqlInsert(tableName string, columns []internal.ColumnDef, value interface{}) (string, []interface{}) {
	key := d.Name() + "://" + tableName
	actual, _ := d.cacheMakeSqlInsert.LoadOrStore(key, &makePostgresSqlInsertInit{})
	init := actual.(*makePostgresSqlInsertInit)
	init.once.Do(func() {
		init.val = d.makeSqlInsert(tableName, columns)
	})
	dataVal := reflect.ValueOf(value)
	if dataVal.Kind() == reflect.Ptr {
		dataVal = dataVal.Elem()
	}
	args := []interface{}{}
	for _, col := range columns {
		if col.IsAuto {
			continue
		}
		fieldVal := dataVal.FieldByName(col.Field.Name)
		if fieldVal.IsValid() {
			args = append(args, fieldVal.Interface())
		} else {
			args = append(args, nil)
		}

	}

	return init.val, args

}
func (d *PostgresDialect) makeSqlInsert(tableName string, columns []internal.ColumnDef) string {

	sql := "INSERT INTO " + d.Quote(tableName) + " ("
	strFields := []string{}
	strValues := []string{}
	i := 1
	RETURNING_ID := ""
	for _, col := range columns {
		if col.IsAuto {
			RETURNING_ID = " RETURNING " + d.Quote(col.Name)
			continue
		}
		strFields = append(strFields, d.Quote(col.Name))
		strValues = append(strValues, d.ToParam(i))
		i++
	}

	sql += strings.Join(strFields, ", ") + ") VALUES (" + strings.Join(strValues, ", ") + ")"
	if RETURNING_ID != "" {
		sql += RETURNING_ID
	}
	return sql
}
