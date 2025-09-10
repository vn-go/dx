package dx

import (
	"database/sql"
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
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
	//sql.db.fecthItems(items, sqlExec.Sql, nil, nil, false, sql.args...)
	return ScanUnsafe(sql.db.DB.DB, sqlExec.Sql, items, sql.args...)
	//return nil
}

type globalOptType struct {
	ShowSql bool
}

var Options = globalOptType{
	ShowSql: false,
}

var metaCache sync.Map // key: reflect.Type.String(), value: map[column]fieldInfo

type fieldInfo struct {
	offset uintptr
	typ    reflect.Type
}

func ScanUnsafe(db *sql.DB, query string, items any, args ...any) error {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("items must be a pointer to slice")
	}

	sliceVal := v.Elem()
	elemType := sliceVal.Type().Elem()

	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	// chuẩn bị map column -> field offset
	fieldMap := make(map[string]struct {
		offset uintptr
		typ    reflect.Type
	})
	for i := 0; i < elemType.NumField(); i++ {
		f := elemType.Field(i)
		name := f.Tag.Get("db")
		if name == "" {
			name = f.Name
		}
		fieldMap[name] = struct {
			offset uintptr
			typ    reflect.Type
		}{f.Offset, f.Type}
	}

	for rows.Next() {
		// tạo struct mới
		newElem := reflect.New(elemType).Elem()
		basePtr := unsafe.Pointer(newElem.UnsafeAddr())

		ptrs := make([]any, len(cols))
		for i, col := range cols {
			if info, ok := fieldMap[col]; ok {
				fieldPtr := unsafe.Add(basePtr, info.offset)
				ptrs[i] = reflect.NewAt(info.typ, fieldPtr).Interface()
			} else {
				// nếu không match field thì scan vào dummy
				var dummy any
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
