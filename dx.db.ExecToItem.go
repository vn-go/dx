package dx

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	dbErrors "github.com/vn-go/dx/errors"

	"github.com/vn-go/dx/model"
)

func (db *DB) ExecToItem(result interface{}, query string, args ...interface{}) error {
	if result == nil {
		return fmt.Errorf("result must not be nil")
	}
	typ := reflect.TypeOf(result)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}
	typ = typ.Elem()
	mapIndex := onTenantDbNeedGetMapIndex(typ)

	return execToItemOptimized(context.Background(), db, result, &mapIndex, query, args...)
}

var onTenantDbNeedGetMapIndexCache sync.Map

type initOnTenantDbNeedGetMapIndex struct {
	once sync.Once
	val  map[string][]int
}

func onTenantDbNeedGetMapIndex(typ reflect.Type) map[string][]int {
	key := typ.String()
	actual, _ := onTenantDbNeedGetMapIndexCache.LoadOrStore(key, &initOnTenantDbNeedGetMapIndex{})
	initBuild := actual.(*initOnTenantDbNeedGetMapIndex)
	initBuild.once.Do(func() {
		initBuild.val = onTenantDbNeedGetMapIndexNoCache(typ)
	})
	return initBuild.val
}
func onTenantDbNeedGetMapIndexNoCache(typ reflect.Type) map[string][]int {
	repoType, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return nil
	}
	ret := map[string][]int{}
	for _, col := range repoType.Entity.Cols {
		ret[col.Field.Name] = col.IndexOfField
	}
	return ret
}

var scanArgsPool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 20)
	},
}

func execToItemOptimized(context context.Context, db *DB, result interface{}, mapIndex *map[string][]int, query string, args ...interface{}) error {
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
		return dbErrors.NewNotFoundErr()
	}

	// Gán lại vào `*result`

	return nil
}

type fieldEncoder struct {
	fieldIndexes [][]int
}
type initGetFieldEncoder struct {
	once sync.Once
	val  fieldEncoder
	err  error
}

var encoderCache sync.Map

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
