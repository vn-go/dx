package mssql

import (
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/vn-go/dx/internal"
)

type makeMssqlSqlInsertInit struct {
	once sync.Once
	val  makeMssqlSqlInsertCacheItem
}
type makeMssqlSqlInsertCacheItem struct {
	sql            string
	indexesOfField [][]int
	fieldNames     []string
	fieldIndex     [][]int
}

func (d *MssqlDialect) MakeSqlInsert(tableName string, columns []internal.ColumnDef, value interface{}) (string, []interface{}) {
	key := d.Name() + "://" + tableName
	actual, _ := d.cacheMakeSqlInsert.LoadOrStore(key, &makeMssqlSqlInsertInit{})
	init := actual.(*makeMssqlSqlInsertInit)
	init.once.Do(func() {
		init.val = d.makeSqlInsert(tableName, columns)
	})
	dataVal := reflect.ValueOf(value)
	if dataVal.Kind() == reflect.Ptr {
		dataVal = dataVal.Elem()
	}
	args := []interface{}{}
	for _, indexOfField := range init.val.indexesOfField {
		fieldVal := dataVal.FieldByIndex(indexOfField)
		if fieldVal.IsValid() {
			args = append(args, fieldVal.Interface())
		} else {
			args = append(args, nil)
		}
	}
	// for _, indexes := range init.val.indexesOfField {

	// 	fieldVal := dataVal.Field(indexes[0])
	// 	for _, i := range indexes[1:] {
	// 		fieldVal = fieldVal.Field(i)
	// 	}
	// 	if fieldVal.IsValid() {
	// 		args = append(args, fieldVal.Interface())
	// 	} else {
	// 		args = append(args, nil)
	// 	}

	// }

	return init.val.sql, args

}
func (d *MssqlDialect) makeSqlInsert(tableName string, columns []internal.ColumnDef) makeMssqlSqlInsertCacheItem {
	sql := "INSERT INTO " + d.Quote(tableName) + " ("
	strFields := []string{}
	strValues := []string{}
	insertedFieldName := ""
	index := 0
	ret := makeMssqlSqlInsertCacheItem{
		sql:            "",
		indexesOfField: [][]int{},
		fieldNames:     []string{},
		fieldIndex:     [][]int{},
	}

	for _, col := range columns {
		if col.IsAuto {
			insertedFieldName = col.Name
			continue
		}
		strFields = append(strFields, d.Quote(col.Name))
		strValues = append(strValues, "@p"+strconv.Itoa(index+1))
		ret.indexesOfField = append(ret.indexesOfField, col.IndexOfField)
		ret.fieldNames = append(ret.fieldNames, col.Field.Name)
		ret.fieldIndex = append(ret.fieldIndex, col.Field.Index)
		index++
	}
	if insertedFieldName != "" {
		sql += strings.Join(strFields, ", ") + ") OUTPUT INSERTED." + d.Quote(insertedFieldName) + " VALUES (" + strings.Join(strValues, ", ") + ")"
	} else {
		sql += strings.Join(strFields, ", ") + ") VALUES (" + strings.Join(strValues, ", ") + ")"
	}
	ret.sql = sql
	return ret
}
