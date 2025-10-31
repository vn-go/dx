package dx

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type structMeta struct {
	fieldPtrs []func(reflect.Value) any // cách lấy con trỏ field
}

var metaCache sync.Map // reflect.Type -> *structMeta

func getStructMeta(t reflect.Type) *structMeta {
	if m, ok := metaCache.Load(t); ok {
		return m.(*structMeta)
	}
	m := &structMeta{fieldPtrs: make([]func(reflect.Value) any, t.NumField())}
	for i := 0; i < t.NumField(); i++ {
		idx := i
		m.fieldPtrs[i] = func(v reflect.Value) any {
			return v.Field(idx).Addr().Interface()
		}
	}
	metaCache.Store(t, m)
	return m
}
func (db *DB) DslQuery(result any, dslQuery string, args ...interface{}) error {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return errors.New("result must be pointer to slice")
	}

	sliceVal := rv.Elem()
	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return errors.New("slice element must be struct")
	}

	// if skip > 0 || limit > 0 {
	// 	dslQuery += ",skip(?),take(?)"
	// 	args = append(args, limit, skip)
	// }

	query, err := db.Smart(dslQuery, args...)
	if err != nil {
		return err
	}
	return db.fecthItems(result, query.Query, nil, nil, true, query.Args...)
	// rows, err := db.Query(query.Query, query.Args...)
	// if err != nil {
	// 	return err
	// }
	// defer rows.Close()

	// meta := getStructMeta(elemType)
	// buf := make([]any, len(meta.fieldPtrs)) // cấp phát 1 lần duy nhất

	// // preallocate slice nếu biết trước limit
	// capHint := 0
	// if limit > 0 {
	// 	capHint = limit
	// }
	// sliceVal = reflect.MakeSlice(sliceVal.Type(), 0, capHint)

	// for rows.Next() {
	// 	elemVal := reflect.New(elemType).Elem()
	// 	for i, f := range meta.fieldPtrs {
	// 		buf[i] = f(elemVal)
	// 	}
	// 	if err := rows.Scan(buf...); err != nil {
	// 		return err
	// 	}
	// 	sliceVal = reflect.Append(sliceVal, elemVal)
	// }

	// rv.Elem().Set(sliceVal)
	// return nil
}

func (db *DB) DslFirstRow(result any, dslQuery string, args ...interface{}) error {
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("result must be pointer to struct")
	}

	elemType := rv.Elem().Type()
	if elemType.Kind() != reflect.Struct {
		return errors.New("struct element must be struct")
	}
	// Compile DSL → SQL
	query, err := db.Smart(dslQuery+",skip(0),take(1)", args...)
	if err != nil {
		return err
	}

	rows, err := db.Query(query.Query, query.Args...)
	if err != nil {
		fmt.Println(query.Query)
		return err
	}
	defer rows.Close()
	meta := getStructMeta(elemType)
	buf := make([]any, len(meta.fieldPtrs)) // cấp phát 1 lần duy nhất

	if rows.Next() {
		elemVal := reflect.New(elemType).Elem()
		for i, f := range meta.fieldPtrs {
			buf[i] = f(elemVal)
		}
		if err := rows.Scan(buf...); err != nil {
			return err
		}
		rv.Elem().Set(elemVal)
	} else {
		return errors.New("no rows found")
	}

	return nil
}
