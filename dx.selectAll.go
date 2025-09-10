package dx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/errors"
	dxErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (db *DB) SelectAll(items any) error {
	// items phải là pointer đến slice
	typ := reflect.TypeOf(items)
	if typ.Kind() != reflect.Ptr {
		return errors.NewSysError(fmt.Sprintf("%s is not pointer to slice", typ.String()))
	}
	sliceVal := reflect.ValueOf(items).Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.NewSysError(fmt.Sprintf("%s is not slice", typ.String()))
	}

	dialect := factory.DialectFactory.Create(db.Info.DriverName)

	// lấy kiểu phần tử của slice
	typElem := sliceVal.Type().Elem()
	if typElem.Kind() == reflect.Ptr {
		typElem = typElem.Elem()
	}

	model, err := model.ModelRegister.GetModelByType(typElem)
	if err != nil {
		return err
	}

	fieldsSelect := make([]string, len(model.Entity.Cols))
	for i, col := range model.Entity.Cols {
		fieldsSelect[i] = dialect.Quote(col.Name) + " AS " + dialect.Quote(col.Field.Name)
	}
	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fieldsSelect, ", "), dialect.Quote(model.Entity.TableName))

	rows, err := db.Query(sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	// chuẩn bị field index/type
	fieldIndexes := make([][]int, len(model.Entity.Cols))
	fieldTypes := make([]reflect.Type, len(model.Entity.Cols))
	for i, col := range model.Entity.Cols {
		fieldIndexes[i] = col.IndexOfField
		fieldTypes[i] = col.Field.Type
	}

	vals := make([]interface{}, len(model.Entity.Cols))
	ptrs := make([]interface{}, len(model.Entity.Cols))
	for i := range ptrs {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}

		// tạo instance struct
		ptr := reflect.New(typElem).Elem()

		for i := range model.Entity.Cols {
			raw := vals[i]
			if raw == nil {
				continue
			}

			val := reflect.ValueOf(raw)
			if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
				continue
			}
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
				if !val.IsValid() {
					continue
				}
			}

			field := ptr.FieldByIndex(fieldIndexes[i])
			if !field.CanSet() {
				continue
			}

			if val.Type().AssignableTo(fieldTypes[i]) {
				field.Set(val)
			} else if val.Type().ConvertibleTo(fieldTypes[i]) {
				field.Set(val.Convert(fieldTypes[i]))
			}
		}

		// append vào slice gốc
		sliceVal.Set(reflect.Append(sliceVal, ptr))
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func (db *DB) SelectAllWithContext(context context.Context, items any) error {
	typ := reflect.TypeOf(items)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Slice {
		return errors.NewSysError(fmt.Sprintf("%s is not slice", reflect.TypeOf(items).String()))
	}
	dialect := factory.DialectFactory.Create(db.Info.DbName)
	model, err := model.ModelRegister.GetModelByType(typ)
	if err != nil {
		return err
	}
	fieldsSelect := make([]string, len(model.Entity.Cols))
	for i, col := range model.Entity.Cols {
		fieldsSelect[i] = dialect.Quote(col.Name) + " AS " + dialect.Quote(col.Field.Name)
	}
	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fieldsSelect, ", "), dialect.Quote(model.Entity.TableName))
	rows, err := db.QueryContext(context, sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	fieldIndexes := make([][]int, len(model.Entity.Cols)) // cache field index paths
	fieldTypes := make([]reflect.Type, len(model.Entity.Cols))
	for i, col := range model.Entity.Cols {
		fieldIndexes[i] = col.IndexOfField
		fieldTypes[i] = col.Field.Type
	}

	// 5. Buffer để scan dữ liệu từ DB
	vals := make([]interface{}, len(model.Entity.Cols))
	ptrs := make([]interface{}, len(model.Entity.Cols))
	for i := range ptrs {
		ptrs[i] = &vals[i]
	}
	valOfItems := reflect.ValueOf(items)
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}

		// Tạo instance T
		ptr := reflect.New(typ).Elem()

		for i := range model.Entity.Cols {
			raw := vals[i]
			if raw == nil {
				continue
			}

			val := reflect.ValueOf(raw)
			if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
				continue
			}
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
				if !val.IsValid() {
					continue
				}
			}

			field := ptr.FieldByIndex(fieldIndexes[i])
			if !field.CanSet() {
				continue
			}

			if val.Type().AssignableTo(fieldTypes[i]) {
				field.Set(val)
			} else if val.Type().ConvertibleTo(fieldTypes[i]) {
				field.Set(val.Convert(fieldTypes[i]))
			}
		}

		reflect.Append(valOfItems, ptr.Addr())
		//items = append(items, ptr.Addr().Interface())
	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
func (db *DB) fecthItems(items any, queryStmt string, ctx context.Context, sqlTx *sql.Tx, resetLen bool, args ...any) error {
	// items phải là pointer đến slice
	typ := reflect.TypeOf(items)
	if typ.Kind() != reflect.Ptr {
		return errors.NewSysError(fmt.Sprintf("%s is not pointer to slice", typ.String()))
	}
	sliceVal := reflect.ValueOf(items).Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.NewSysError(fmt.Sprintf("%s is not slice", typ.String()))
	}
	if resetLen {
		sliceVal.SetLen(0)
	}

	// lấy kiểu phần tử của slice
	typElem := sliceVal.Type().Elem()
	if typElem.Kind() == reflect.Ptr {
		typElem = typElem.Elem()
	}
	var rows *sql.Rows
	var err error
	if sqlTx != nil {
		stmt, err := sqlTx.Prepare(queryStmt)
		if err != nil {
			return dxErrors.NewSqlExecError(
				"Exec error", queryStmt, db.DriverName, err,
			)
		}
		rows, err = stmt.Query(args...)
		if err != nil {
			return dxErrors.NewSqlExecError(
				"Exec error", queryStmt, db.DriverName, err,
			)
		}
	} else {
		stmt, err := db.Prepare(queryStmt)
		if err != nil {
			return dxErrors.NewSqlExecError(
				"Exec error", queryStmt, db.DriverName, err,
			)
		}
		if ctx != nil {
			rows, err = stmt.QueryContext(ctx, args...)
			if err != nil {
				return dxErrors.NewSqlExecError(
					"Exec error", queryStmt, db.DriverName, err,
				)
			}
		} else {
			rows, err = stmt.Query(args...)
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
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {

		return db.parseError(err)
	}

	//vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	// for i := range ptrs {
	// 	ptrs[i] = &vals[i]
	// }
	key := typElem.String() + "://" + queryStmt + "://fecthItems"
	fectInfo, err := internal.OnceCall(key, func() ([]struct {
		fieldIndexes []int
		fieldType    reflect.Type
	}, error) {
		ret := make([]struct {
			fieldIndexes []int
			fieldType    reflect.Type
		}, len(cols))
		for i, col := range cols {
			if field, ok := typElem.FieldByNameFunc(func(s string) bool {
				r := []rune(s)
				return unicode.IsUpper(r[0]) && strings.EqualFold(s, col)
			}); ok {
				ret[i].fieldIndexes = field.Index
				ret[i].fieldType = field.Type
			}
		}
		return ret, nil
	})
	if err != nil {
		return err
	}
	for rows.Next() {
		ptr := reflect.New(typElem).Elem()
		for i, fx := range fectInfo {
			ptrs[i] = ptr.FieldByIndex(fx.fieldIndexes).Addr().Interface()
		}

		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		//data := ptr.Interface()
		sliceVal.Set(reflect.Append(sliceVal, ptr))

	}

	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}
func (db *DB) fecthItem(
	item any,
	queryStmt string,
	ctx context.Context, sqlTx *sql.Tx, resetLen bool, args ...any) error {
	if Options.ShowSql {
		fmt.Println(queryStmt)
	}
	// items phải là pointer đến slice
	typ := reflect.TypeOf(item)

	// lấy kiểu phần tử của slice
	typElem := typ
	if typElem.Kind() == reflect.Ptr {
		typElem = typElem.Elem()
	}
	var rows *sql.Rows
	var err error
	if sqlTx != nil {
		stmt, err := sqlTx.Prepare(queryStmt)
		if err != nil {
			return err
		}
		rows, err = stmt.Query(args...)
		if err != nil {
			return err
		}
	} else {
		stmt, err := db.Prepare(queryStmt)
		if err != nil {
			return err
		}
		if ctx != nil {
			rows, err = stmt.QueryContext(ctx, args...)
			if err != nil {
				return err
			}
		} else {
			rows, err = stmt.Query(args...)
			if err != nil {
				return err
			}
		}

	}

	if err != nil {
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {

		return db.parseError(err)
	}

	//vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	// for i := range ptrs {
	// 	ptrs[i] = &vals[i]
	// }
	key := typElem.String() + "://" + queryStmt + "://fecthItem"
	fectInfo, err := internal.OnceCall(key, func() ([]struct {
		fieldIndexes []int
		fieldType    reflect.Type
	}, error) {
		ret := make([]struct {
			fieldIndexes []int
			fieldType    reflect.Type
		}, len(cols))
		for i, col := range cols {
			if field, ok := typElem.FieldByNameFunc(func(s string) bool {
				r := []rune(s)
				return unicode.IsUpper(r[0]) && strings.EqualFold(s, col)
			}); ok {
				ret[i].fieldIndexes = field.Index
				ret[i].fieldType = field.Type
			}
		}
		return ret, nil
	})
	if err != nil {
		return err
	}
	ptr := reflect.ValueOf(item).Elem()
	for i, fx := range fectInfo {
		ptrs[i] = ptr.FieldByIndex(fx.fieldIndexes).Addr().Interface()
	}
	found := false
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		found = true
	}

	if err := rows.Err(); err != nil {
		return err
	}
	if !found {
		return dxErrors.NewNotFoundErr()
	}
	return nil
}
