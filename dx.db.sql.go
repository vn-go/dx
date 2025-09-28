package dx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unsafe"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
	dxErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"
)

type sqlStatementType struct {
	sql  string
	args []any
	db   *DB
}

func (db *DB) Sql(sqlStatement string, args ...any) *sqlStatementType {
	return &sqlStatementType{
		sql:  sqlStatement,
		args: args,
		db:   db,
	}
}
func (sql *sqlStatementType) GetExecSql() (*types.SqlParse, error) {
	key := "sqlStatementType/GetExecSql/" + sql.sql
	return internal.OnceCall(key, func() (*types.SqlParse, error) {
		info, err := compiler.Compile(sql.sql, sql.db.DriverName)
		if err != nil {
			return nil, err
		}

		return factory.DialectFactory.Create(sql.db.DriverName).BuildSql(info)
	})

}
func (sql *sqlStatementType) ScanRow(items any) error {
	if err := internal.Helper.AddrssertSinglePointerToSlice(items); err != nil {
		return err
	}
	sqlExec, err := sql.GetExecSql()
	if err != nil {
		return err
	}
	err = sql.db.fecthItems(items, sqlExec.Sql, nil, nil, false, sql.args...)
	if err != nil {
		return sql.db.parseError(err)
	}
	return nil
}

type globalOptType struct {
	ShowSql bool
}

var Options = globalOptType{
	ShowSql: false,
}

func (db *DB) fecthItems(items any, queryStmt string, ctx context.Context, sqlTx *sql.Tx, resetLen bool, args ...any) error {
	// items phải là pointer đến slice
	if err := internal.Helper.AddrssertSinglePointerToSlice(items); err != nil {
		return err
	}
	typ := reflect.TypeOf(items)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sliceVal := reflect.ValueOf(items).Elem()
	return db.fecthItemsBySliceVal(sliceVal, queryStmt, ctx, sqlTx, resetLen, args...)

}
func (db *DB) fecthItemsBySliceVal(sliceVal reflect.Value, queryStmt string, ctx context.Context, sqlTx *sql.Tx, resetLen bool, args ...any) error {
	typ := sliceVal.Type()
	if resetLen {
		sliceVal.SetLen(0)
	}

	// // lấy kiểu phần tử của slice
	typElem := typ.Elem()
	if typElem.Kind() == reflect.Ptr {
		typElem = typElem.Elem()
	}
	var rows *sql.Rows
	var err error
	if sqlTx != nil {
		if ctx == nil {
			rows, err = sqlTx.Query(queryStmt, args...)
			if err != nil {
				return dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		} else {
			rows, err = sqlTx.QueryContext(ctx, queryStmt, args...)
			if err != nil {
				return dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		}

	} else {
		//stmt, err := db.Prepare(queryStmt)
		if err != nil {
			return dxErrors.NewSqlExecError(
				"Exec error", queryStmt, db.DriverName, err,
			)
		}
		if ctx != nil {
			rows, err = db.QueryContext(ctx, queryStmt, args...)
			//rows, err = stmt.QueryContext(ctx, args...)
			if err != nil {
				return dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		} else {
			rows, err = db.Query(queryStmt, args...)
			if err != nil {
				return dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		}

	}

	if err != nil {
		return err
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
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
		return err
	}
	if sliceVal.CanAddr() {
		return fetchUnsafe(rows, sliceVal.Addr().Interface(), cols, fectInfo)
	} else {
		return fetchUnsafeValue(rows, sliceVal, cols, fectInfo)
	}

}
func (db *DB) fecthItemsOfType(typ reflect.Type, queryStmt string, ctx context.Context, sqlTx *sql.Tx, resetLen bool, args ...any) (reflect.Value, error) {
	sliceType := reflect.SliceOf(typ)

	sliceTypeValPtr := reflect.New(sliceType)
	err := db.fecthItemsBySliceVal(sliceTypeValPtr.Elem(), queryStmt, ctx, nil, false, args...)
	if err != nil {
		return reflect.Value{}, err
	}

	return sliceTypeValPtr.Elem().Addr(), nil

}

type fieldInfo struct {
	offset uintptr
	typ    reflect.Type
}

func fetchUnsafe(rows *sql.Rows, items any, cols []string, fieldMap map[string]fieldInfo) error {
	defer rows.Close()
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("items must be a pointer to slice")
	}

	sliceVal := v.Elem()
	elemType := sliceVal.Type().Elem()

	ptrs := make([]any, len(cols))
	var dummy interface{}

	for rows.Next() {
		newElem := reflect.New(elemType).Elem()
		basePtr := unsafe.Pointer(newElem.UnsafeAddr())

		for i, col := range cols {
			if info, ok := fieldMap[col]; ok {
				fieldPtr := unsafe.Add(basePtr, info.offset)
				ptrs[i] = reflect.NewAt(info.typ, fieldPtr).Interface()
			} else {
				ptrs[i] = &dummy
			}
		}

		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, newElem))
	}

	return rows.Err()
}
func fetchUnsafeValue(rows *sql.Rows, sliceVal reflect.Value, cols []string, fieldMap map[string]fieldInfo) error {
	defer rows.Close()

	elemType := sliceVal.Type().Elem()

	ptrs := make([]any, len(cols))
	var dummy interface{}

	for rows.Next() {
		newElem := reflect.New(elemType).Elem()
		basePtr := unsafe.Pointer(newElem.UnsafeAddr())

		for i, col := range cols {
			if info, ok := fieldMap[col]; ok {
				fieldPtr := unsafe.Add(basePtr, info.offset)
				ptrs[i] = reflect.NewAt(info.typ, fieldPtr).Interface()
			} else {
				ptrs[i] = &dummy
			}
		}

		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		sliceVal.Set(reflect.Append(sliceVal, newElem))
	}

	return rows.Err()
}
