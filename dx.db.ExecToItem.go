package dx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	dbErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"

	"github.com/vn-go/dx/model"
)

func (db *DB) ExecToItem(result interface{}, query string, ctx context.Context, sqlTx *sql.Tx, args ...interface{}) error {
	if result == nil {
		return fmt.Errorf("result must not be nil")
	}
	typ := reflect.TypeOf(result)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}
	typ = typ.Elem()
	key := typ.String() + "://" + db.DriverName + "/ExecToItem/" + query
	ret, err := internal.OnceCall(key, func() (*map[string][]int, error) {
		repoType, err := model.ModelRegister.GetModelByType(typ)
		if err != nil {
			return nil, err
		}
		ret := map[string][]int{}
		for _, col := range repoType.Entity.Cols {
			ret[col.Field.Name] = col.IndexOfField
		}
		return &ret, nil
	})
	if err != nil {
		return err
	}
	//mapIndex := onTenantDbNeedGetMapIndex(typ)
	if ctx == nil {
		ctx = context.Background()
	}
	return db.execToItemOptimized(ctx, sqlTx, result, ret, query, args...)
}

var scanArgsPool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 20)
	},
}

func (db *DB) execToItemOptimized(context context.Context, sqlTx *sql.Tx, result interface{}, mapIndex *map[string][]int, query string, args ...interface{}) error {
	if Options.ShowSql {
		fmt.Println(query)
	}
	ptrVal := reflect.ValueOf(result)
	if ptrVal.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to slice")
	}

	typ := reflect.TypeOf(result)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("result must be a pointer to struct")
	}
	typ = typ.Elem()

	var rows *sql.Rows
	var err error
	if sqlTx != nil {
		stm, err := sqlTx.Prepare(query)
		if err != nil {
			return err
		}
		rows, err = stm.QueryContext(context, args...)
		if err != nil {
			return err
		}
	} else {
		stm, err := db.DB.Prepare(query)
		if err != nil {
			return err
		}
		rows, err = stm.QueryContext(context, args...)
		if err != nil {
			return err
		}
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	fieldIndexes, err := db.getFieldEncoder(typ, cols, mapIndex)
	if err != nil {
		return err
	}
	row := reflect.ValueOf(result).Elem()

	rowCount := 0

	for rows.Next() {

		scanArgs := scanArgsPool.Get().([]interface{})[:0]
		for _, idx := range fieldIndexes {
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

func (db *DB) getFieldEncoder(typ reflect.Type, cols []string, mapIndex *map[string][]int) ([][]int, error) {
	key := typ.String() + "://" + strings.Join(cols, ",")
	return internal.OnceCall(key, func() ([][]int, error) {
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

		return fields, nil
	})

}
