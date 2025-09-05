package dbutils

import (
	"fmt"
	"reflect"

	"github.com/vn-go/dx/entity"
)

/*
Rac roi o cho muon lay gia tri cua cac cot tự tăng khi insert trong sql server
phai dung db.Queryrow(sql, args)
Vi vay ham nay chi dung voi sqlserver
------------------------
There is confusion about how to get the value of auto-increment columns when inserting in SQL Server.
You must use db.QueryRow(sql, args).
Therefore, this function only works with SQL Server.
*/
func (r *inserter) fetchAfterInsertForQueryRow(entity *entity.Entity, dataValue reflect.Value, insertedValue any) error {
	autoCols := entity.GetAutoValueColumns()
	if len(autoCols) == 0 || insertedValue == nil {
		return nil
	}
	if len(autoCols) != 1 {
		return fmt.Errorf("only single auto-increment column is supported for QueryRow insert")
	}

	col := autoCols[0]
	field := dataValue.FieldByName(col.Field.Name)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("cannot set field %s", col.Field.Name)
	}

	val := reflect.ValueOf(insertedValue)
	// Nếu là ptr như *int64
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	// Gán giá trị nếu kiểu phù hợp
	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
	} else if val.Type().ConvertibleTo(field.Type()) {
		field.Set(val.Convert(field.Type()))
	} else {
		return fmt.Errorf("cannot assign insert id type %s to field %s type %s",
			val.Type(), col.Field.Name, field.Type())
	}

	return nil
}
