package dx

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/model"
)

func (db *DB) SelectAll(items any) error {
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
	rows, err := db.Query(sql)
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
