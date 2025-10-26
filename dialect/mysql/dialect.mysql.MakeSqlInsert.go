package mysql

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
)

type makeMySqlSqlInsertInit struct {
	once sync.Once
	val  string
}

func (d *mySqlDialect) MakeSqlInsert(tableName string, columns []entity.ColumnDef, value interface{}) (string, []interface{}) {

	sql, cols := d.makeSqlInsert(tableName, columns)
	dataVal := reflect.ValueOf(value)
	if dataVal.Kind() == reflect.Ptr {
		dataVal = dataVal.Elem()
	}

	args := make([]interface{}, len(cols))
	for i, col := range cols {
		if col.IsAuto {
			continue
		}
		fieldVal := dataVal.FieldByName(col.Field.Name)
		if fieldVal.IsValid() {
			args[i] = fieldVal.Interface() //append(args, fieldVal.Interface())
			typ := col.Field.Type
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			if typ.Kind() == reflect.Slice {
				bff, err := json.Marshal(args[i])
				if err == nil {
					args[i] = bff
				}
			}
		}

	}

	return sql, args

}

func (d *mySqlDialect) makeSqlInsert(tableName string, columns []entity.ColumnDef) (string, []entity.ColumnDef) {
	key := "mySqlDialect/makeSqlInsert" + "@" + tableName

	ret, _ := internal.OnceCall(key, func() (struct {
		sql string

		columns []entity.ColumnDef
	}, error) {
		sql := "INSERT INTO " + d.Quote(tableName) + " ("
		strFields := []string{}
		strValues := []string{}
		numOfColumn := 0
		cols := []entity.ColumnDef{}
		for _, col := range columns {
			if col.IsAuto {
				// MySQL: bỏ qua trường tự tăng, nhưng không dùng OUTPUT như SQL Server
				continue
			}
			strFields = append(strFields, d.Quote(col.Name))
			strValues = append(strValues, "?")
			numOfColumn++
			cols = append(cols, col)
		}

		sql += strings.Join(strFields, ", ") + ") VALUES (" + strings.Join(strValues, ", ") + ")"
		ret := struct {
			sql     string
			columns []entity.ColumnDef
		}{
			sql:     sql,
			columns: cols,
		}
		return ret, nil
	})
	return ret.sql, ret.columns

}
