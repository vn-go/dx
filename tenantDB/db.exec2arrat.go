package tenantDB

import (
	"context"
	"database/sql"

	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

type onGetMapIndex func(typ reflect.Type) map[string][]int

var OnGetMapIndex onGetMapIndex

func (db *TenantDB) ExecToItemWithContext(context context.Context, result interface{}, query string, args ...interface{}) error {
	if result == nil {
		return fmt.Errorf("result must not be nil")
	}
	typ := reflect.TypeOf(result)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}
	typ = typ.Elem()
	mapIndex := OnGetMapIndex(typ)

	return execToItemOptimized(context, db, result, &mapIndex, query, args...)
}
func (db *TenantDB) ExecToItem(result interface{}, query string, args ...interface{}) error {
	if result == nil {
		return fmt.Errorf("result must not be nil")
	}
	typ := reflect.TypeOf(result)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}
	typ = typ.Elem()
	mapIndex := OnGetMapIndex(typ)

	return execToItemOptimized(context.Background(), db, result, &mapIndex, query, args...)
}
func (db *TenantDB) ExecToArray(result interface{}, query string, args ...interface{}) error {
	if result == nil {
		return fmt.Errorf("result must not be nil")
	}

	typ := reflect.TypeOf(result)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to slice")
	}
	start := time.Now()
	ret := execToArrayOptimized(db, result, query, args...)
	//ret := exec2array(db, result, query, args...)
	n := time.Since(start).Milliseconds()
	fmt.Println(n)
	return ret
}
func execToArray_original(db *TenantDB, typ reflect.Type, query string, args ...interface{}) ([]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []interface{}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		row := reflect.New(typ).Interface()
		scanArgs := make([]interface{}, len(cols))
		for i, col := range cols {
			scanArgs[i] = reflect.ValueOf(row).Elem().FieldByName(col).Addr().Interface()
		}
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil

}
func execToArrayFromChatGPT_V1(db *TenantDB, typ reflect.Type, query string, args ...interface{}) ([]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Precompute: field index for each column name
	fieldIndexes := make([][]int, len(cols))
	for i, colName := range cols {
		field, ok := typ.FieldByName(colName)
		if !ok {
			return nil, fmt.Errorf("column '%s' not found in struct %s", colName, typ.Name())
		}
		fieldIndexes[i] = field.Index
	}

	// Result list
	var result []interface{}

	for rows.Next() {
		// Create new struct
		rowVal := reflect.New(typ).Elem()

		// Prepare scan targets
		scanArgs := make([]interface{}, len(cols))
		for i, fieldIdx := range fieldIndexes {
			field := rowVal.FieldByIndex(fieldIdx)
			scanArgs[i] = field.Addr().Interface()
		}

		// Scan row into struct
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		result = append(result, rowVal.Addr().Interface())
	}

	return result, nil
}

type fieldEncoder struct {
	fieldIndexes [][]int
}

var encoderCache sync.Map // map[reflect.Type]*fieldEncoder

var scanArgsPool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 20)
	},
}

var rowValPool sync.Pool // per-type struct pool
type initGetFieldEncoder struct {
	once sync.Once
	val  fieldEncoder
	err  error
}

func getFieldEncoder(typ reflect.Type, cols []string, mapIndex *map[string][]int) (*fieldEncoder, error) {
	key := typ.String() + "://" + strings.Join(cols, ",")
	actual, _ := encoderCache.LoadOrStore(key, &initGetFieldEncoder{})
	init := actual.(*initGetFieldEncoder)
	init.once.Do(func() {
		val, err := getFieldEncoderNoCache(typ, cols, mapIndex)
		init.val = *val
		init.err = err
	})
	return &init.val, init.err

}
func getFieldEncoderNoCache(typ reflect.Type, cols []string, mapIndex *map[string][]int) (*fieldEncoder, error) {

	fields := make([][]int, len(cols))
	for i, col := range cols {
		// Try exact match first
		field, ok := typ.FieldByName(col)
		if !ok {
			// Try case-insensitive match
			for j := 0; j < typ.NumField(); j++ {
				if strings.EqualFold(typ.Field(j).Name, col) {
					field = typ.Field(j)
					ok = true
					break
				}
			}
		}
		if !ok {
			return nil, fmt.Errorf("column %s not found in struct", col)
		}
		if mapIndex == nil {
			fields[i] = field.Index
		} else {

			fields[i] = (*mapIndex)[field.Name]
		}
	}

	encoder := &fieldEncoder{fieldIndexes: fields}

	return encoder, nil
}

/*
	 Example usage:
		type MyStruct struct {
			ID   int
		}
		items:=[]MyStruct{}
		execToArrayOptimized(db,&result,"SELECT * FROM my_table", args...)
*/
func execToArrayOptimized(db *TenantDB, result interface{}, query string, args ...interface{}) error {
	ptrVal := reflect.ValueOf(result)
	if ptrVal.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to slice")
	}

	sliceVal := ptrVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice")
	}

	elemType := sliceVal.Type().Elem() // MyStruct hoặc *MyStruct
	typ := elemType
	if elemType.Kind() == reflect.Ptr {
		typ = elemType.Elem()
	}
	stm, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	rows, err := stm.Query(args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	encoder, err := getFieldEncoder(typ, cols, nil)
	if err != nil {
		return err
	}

	newSlice := reflect.MakeSlice(sliceVal.Type(), 0, 0)

	for rows.Next() {
		var row reflect.Value

		val := rowValPool.Get()
		if val == nil {
			row = reflect.New(typ)
		} else {
			row = reflect.ValueOf(val)
		}

		rowVal := row.Elem()

		scanArgs := scanArgsPool.Get().([]interface{})[:0]
		for _, idx := range encoder.fieldIndexes {
			scanArgs = append(scanArgs, rowVal.FieldByIndex(idx).Addr().Interface())
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}

		// Gán row vào slice
		var finalVal reflect.Value
		if elemType.Kind() == reflect.Ptr {
			finalVal = row
		} else {
			finalVal = row.Elem()
		}
		newSlice = reflect.Append(newSlice, finalVal)

		rowValPool.Put(row.Interface())
		scanArgsPool.Put(scanArgs)
	}

	// Gán lại vào `*result`
	sliceVal.Set(newSlice)
	return nil
}

type execToItemOptimizedErrorNotFound func() error

var ExecToItemOptimizedErrorNotFound execToItemOptimizedErrorNotFound

func execToItemOptimized(context context.Context, db *TenantDB, result interface{}, mapIndex *map[string][]int, query string, args ...interface{}) error {
	ptrVal := reflect.ValueOf(result)
	if ptrVal.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to slice")
	}

	typ := reflect.TypeOf(result)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}
	typ = typ.Elem()

	stm, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	rows, err := stm.QueryContext(context, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	encoder, err := getFieldEncoder(typ, cols, mapIndex)
	if err != nil {
		return err
	}
	row := reflect.ValueOf(result).Elem()

	rowCount := 0

	for rows.Next() {

		scanArgs := scanArgsPool.Get().([]interface{})[:0]
		for _, idx := range encoder.fieldIndexes {
			scanArgs = append(scanArgs, row.FieldByIndex(idx).Addr().Interface())
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}
		rowCount++
		// Gán row vào slice

	}
	if rowCount == 0 {
		return ExecToItemOptimizedErrorNotFound()
	}

	// Gán lại vào `*result`

	return nil
}

type exec2arrayFn func(db *TenantDB, result interface{}, query string, args ...interface{}) error

var exec2array exec2arrayFn = execToArrayOptimized

type FieldDecoder struct {
	Index    int                                        // field index trong struct
	ScanType reflect.Type                               // type dùng trong rows.Scan (ex: *sql.NullString)
	SetValue func(field reflect.Value, val interface{}) // logic convert từ scan value → gán vào struct
}

var decoderCache sync.Map // map[reflect.Type][]FieldDecoder
func getOrBuildFieldDecoders(typ reflect.Type) []FieldDecoder {
	if cached, ok := decoderCache.Load(typ); ok {
		return cached.([]FieldDecoder)
	}

	var decoders []FieldDecoder
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		dec := FieldDecoder{
			Index: i,
		}

		switch field.Type.Kind() {
		case reflect.String:
			dec.ScanType = reflect.TypeOf(new(sql.NullString)).Elem()
			dec.SetValue = func(f reflect.Value, v interface{}) {
				val := v.(sql.NullString)
				if val.Valid {
					f.SetString(val.String)
				} else {
					f.SetString("")
				}
			}
		case reflect.Int, reflect.Int64, reflect.Int32:
			dec.ScanType = reflect.TypeOf(new(sql.NullInt64)).Elem()
			dec.SetValue = func(f reflect.Value, v interface{}) {
				val := v.(sql.NullInt64)
				if val.Valid {
					f.SetInt(val.Int64)
				} else {
					f.SetInt(0)
				}
			}
		case reflect.Float64:
			dec.ScanType = reflect.TypeOf(new(sql.NullFloat64)).Elem()
			dec.SetValue = func(f reflect.Value, v interface{}) {
				val := v.(sql.NullFloat64)
				if val.Valid {
					f.SetFloat(val.Float64)
				} else {
					f.SetFloat(0)
				}
			}
		case reflect.Bool:
			dec.ScanType = reflect.TypeOf(new(sql.NullBool)).Elem()
			dec.SetValue = func(f reflect.Value, v interface{}) {
				val := v.(sql.NullBool)
				if val.Valid {
					f.SetBool(val.Bool)
				} else {
					f.SetBool(false)
				}
			}
		case reflect.Struct:
			if field.Type == reflect.TypeOf(time.Time{}) {
				dec.ScanType = reflect.TypeOf(new(sql.NullTime)).Elem()
				dec.SetValue = func(f reflect.Value, v interface{}) {
					val := v.(sql.NullTime)
					if val.Valid {
						f.Set(reflect.ValueOf(val.Time))
					} else {
						f.Set(reflect.Zero(f.Type()))
					}
				}
			}
		default:
			continue // hoặc fallback
		}
		decoders = append(decoders, dec)
	}
	decoderCache.Store(typ, decoders)
	return decoders
}

var scanPool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 32)
	},
}

func execToArraySafe(db *TenantDB, typ reflect.Type, query string, args ...any) ([]interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decoders := getOrBuildFieldDecoders(typ)
	n := len(decoders)

	poolItem := scanPool.Get().([]interface{})
	if cap(poolItem) < n {
		poolItem = make([]interface{}, n)
	} else {
		poolItem = poolItem[:n]
	}
	defer scanPool.Put(poolItem)

	var result []interface{}

	for rows.Next() {
		rowVal := reflect.New(typ).Elem()
		for i, dec := range decoders {
			ptr := reflect.New(dec.ScanType)
			poolItem[i] = ptr.Interface()
		}

		if err := rows.Scan(poolItem[:n]...); err != nil {
			return nil, err
		}

		for i, dec := range decoders {
			val := reflect.ValueOf(poolItem[i]).Elem().Interface()
			field := rowVal.Field(dec.Index)
			dec.SetValue(field, val)
		}

		result = append(result, rowVal.Addr().Interface())
	}

	return result, nil
}
