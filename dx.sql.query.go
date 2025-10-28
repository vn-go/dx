package dx

import (
	"errors"
	"reflect"
)

func newRowPtrFromStruct(t reflect.Type) ([]any, reflect.Value) {
	v := reflect.New(t).Elem()          // tạo struct instance
	fields := make([]any, t.NumField()) // chứa địa chỉ field để Scan
	for i := 0; i < t.NumField(); i++ {
		fields[i] = v.Field(i).Addr().Interface() // truyền con trỏ đến từng field
	}
	return fields, v
}

func (db *DB) DslQuery(result any, dslQuery string, skip, limit int, args ...interface{}) error {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return errors.New("result must be pointer to slice")
	}

	sliceVal := rv.Elem()
	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errors.New("slice element must be struct")
	}
	if skip > 0 || limit > 0 {
		dslQuery = dslQuery + ",skip(?),take(?)"
		args = append(args, limit, skip)
	}
	// Compile DSL → SQL
	query, err := db.Smart(dslQuery, args...)
	if err != nil {
		return err
	}

	rows, err := db.Query(query.Query, query.Args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		fieldPtrs, elemVal := newRowPtrFromStruct(elemType)
		if err := rows.Scan(fieldPtrs...); err != nil {
			return err
		}
		sliceVal = reflect.Append(sliceVal, elemVal)
	}

	rv.Elem().Set(sliceVal)
	return nil
}
