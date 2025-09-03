package tenantDB

import (
	"database/sql"
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// FieldInfo dùng để lưu metadata cho mỗi field
type fieldInfo struct {
	offset uintptr
	typ    reflect.Type
}

// entityInfo chứa metadata của 1 struct
type entityInfo struct {
	fields   []*fieldInfo
	newValue func() reflect.Value // nhanh hơn reflect.New
}

var (
	entityCache sync.Map // map[reflect.Type]*entityInfo
)

type initGetEntityInfo struct {
	once sync.Once
	info *entityInfo
}

var getEntityInfoCache = sync.Map{} // map[reflect.Type]*initGetEntityInfo
func getEntityInfo(t reflect.Type) *entityInfo {
	actual, _ := getEntityInfoCache.LoadOrStore(t, &initGetEntityInfo{})
	initInfo := actual.(*initGetEntityInfo)
	initInfo.once.Do(func() {
		initInfo.info = getEntityInfoNoCache(t)
	})
	return initInfo.info

}

// getEntityInfo tạo thông tin field mapping một lần duy nhất
func getEntityInfoNoCache(t reflect.Type) *entityInfo {

	//ptrType := reflect.PtrTo(t)

	numField := t.NumField()
	fields := []*fieldInfo{}

	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // unexported field
			fields = append(fields, nil)
			continue
		}
		fields = append(fields, &fieldInfo{
			offset: f.Offset,
			typ:    f.Type,
		})
	}

	info := &entityInfo{
		fields:   fields,
		newValue: func() reflect.Value { return reflect.New(t).Elem() },
	}

	return info
}

func ToArrayUnsafeFast(rows *sql.Rows, dest interface{}) error {
	defer rows.Close()

	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	sliceVal := destVal.Elem()
	elemType := sliceVal.Type().Elem()

	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("only slice of struct is supported")
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	numCols := len(cols)

	entity := getEntityInfo(elemType)
	scanArgs := make([]interface{}, numCols)

	for rows.Next() {
		obj := reflect.New(elemType).Elem()
		// #nosec G103 -- using unsafe.Pointer with reflect.UnsafeAddr for zero-copy field mapping
		base := unsafe.Pointer(obj.UnsafeAddr())

		for i := 0; i < numCols; i++ {
			if i >= len(entity.fields) || entity.fields[i] == nil {
				var dummy interface{}
				scanArgs[i] = &dummy
				continue
			}

			fi := entity.fields[i]
			// #nosec G103 -- using unsafe.Pointer with reflect.UnsafeAddr for zero-copy field mapping
			fieldPtr := unsafe.Pointer(uintptr(base) + fi.offset)
			scanArgs[i] = reflect.NewAt(fi.typ, fieldPtr).Interface()
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, obj))
	}

	return rows.Err()
}
