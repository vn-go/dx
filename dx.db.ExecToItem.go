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

func (db *DB) execToItemOptimized(
	ctx context.Context,
	sqlTx *sql.Tx,
	result interface{},
	mapIndex *map[string][]int,
	query string,
	args ...interface{},
) (err error) { // dùng named return

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
	var stm *sql.Stmt

	if sqlTx != nil {
		stm, err = sqlTx.Prepare(query)
		if err != nil {
			return
		}
	} else {
		stm, err = db.DB.Prepare(query)
		if err != nil {
			return
		}
	}

	defer func() {
		if cerr := stm.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	rows, err = stm.QueryContext(ctx, args...)
	if err != nil {
		return
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	cols, err := rows.Columns()
	if err != nil {
		return
	}

	fieldIndexes, err := db.getFieldEncoder(typ, cols, mapIndex)
	if err != nil {
		return
	}

	row := reflect.ValueOf(result).Elem()
	rowCount := 0

	for rows.Next() {
		scanArgs := scanArgsPool.Get().([]interface{})[:0]
		for _, idx := range fieldIndexes {
			scanArgs = append(scanArgs, row.FieldByIndex(idx).Addr().Interface())
		}

		if err = rows.Scan(scanArgs...); err != nil {
			return
		}
		rowCount++
	}

	if rowCount == 0 {
		return dbErrors.NewNotFoundErr()
	}

	return
}

type getFieldEncoderKey struct {
	typ  reflect.Type
	cols []string
}

func (db *DB) getFieldEncoder(typ reflect.Type, cols []string, mapIndex *map[string][]int) ([][]int, error) {
	//key := typ.String() + "://" + strings.Join(cols, ",")
	key := getFieldEncoderKey{
		typ:  typ,
		cols: cols,
	}
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
