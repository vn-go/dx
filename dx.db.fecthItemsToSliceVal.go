package dx

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"unicode"
	"unsafe"

	dxErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"
)

func (db *DB) fecthItemsToSliceVal(queryStmt string, ctx context.Context, sqlTx *sql.Tx, resetLen bool, args ...any) (any, error) {

	var rows *sql.Rows
	var err error
	if db.DriverName == "mysql" {
		queryStmt, args, err = internal.Helper.FixParam(queryStmt, args)
		if err != nil {
			return nil, err
		}
	}
	if Options.ShowSql {
		println("----------------------------")
		println(queryStmt)
		println("----------------------------")
	}
	if sqlTx != nil {
		if ctx == nil {
			rows, err = sqlTx.Query(queryStmt, args...)
			if err != nil {
				return nil, dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		} else {
			rows, err = sqlTx.QueryContext(ctx, queryStmt, args...)
			if err != nil {
				return nil, dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		}

	} else {
		//stmt, err := db.Prepare(queryStmt)
		if err != nil {
			return nil, dxErrors.NewSqlExecError(
				"Exec error", queryStmt, db.DriverName, err,
			)
		}
		if ctx != nil {
			rows, err = db.QueryContext(ctx, queryStmt, args...)
			//rows, err = stmt.QueryContext(ctx, args...)
			if err != nil {
				return nil, dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		} else {
			rows, err = db.Query(queryStmt, args...)
			if err != nil {
				return nil, dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		}

	}

	if err != nil {
		return nil, err
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	typElem := internal.Helper.CreateDynamicStructFromSqlColumnType(colTypes)
	sliceVal := reflect.New(reflect.SliceOf(typElem))
	key := typElem.String() + "://" + queryStmt + "://fecthItems"
	fectInfo, err := internal.OnceCall(key, func() (map[string]fieldInfo, error) {
		ret := make(map[string]fieldInfo)
		for _, col := range cols {
			if field, ok := typElem.FieldByNameFunc(func(s string) bool {
				r := []rune(s)
				return unicode.IsUpper(r[0]) && strings.EqualFold(s, col)
			}); ok {
				ret[field.Name] = fieldInfo{
					offset: field.Offset,
					typ:    field.Type,
				}
			}
		}
		return ret, nil
	})
	if err != nil {
		return nil, err
	}
	var dummy interface{}
	for rows.Next() {
		newElem := reflect.New(typElem).Elem()
		basePtr := unsafe.Pointer(newElem.UnsafeAddr())
		ptrs := make([]any, len(colTypes))
		for i, col := range cols {
			if info, ok := fectInfo[col]; ok {
				fieldPtr := unsafe.Add(basePtr, info.offset)
				ptrs[i] = reflect.NewAt(info.typ, fieldPtr).Interface()
			} else {
				ptrs[i] = &dummy
			}
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}

		// ✅ fix here
		sliceVal.Elem().Set(reflect.Append(sliceVal.Elem(), newElem))
	}

	return sliceVal.Interface(), nil
}
