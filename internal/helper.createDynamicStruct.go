package internal

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// ---- convert sql type → Go type ----
func (h *helperType) goTypeFromSqlColumn(col *sql.ColumnType) reflect.Type {
	// nếu driver hỗ trợ thông tin chính xác
	if scanType := col.ScanType(); scanType != nil {
		return scanType
	}

	dbType := strings.ToLower(col.DatabaseTypeName())

	switch {
	case strings.Contains(dbType, "int"):
		return reflect.TypeFor[int64]()
	case strings.Contains(dbType, "bool"):
		return reflect.TypeOf(bool(false))
	case strings.Contains(dbType, "float"), strings.Contains(dbType, "double"), strings.Contains(dbType, "numeric"), strings.Contains(dbType, "decimal"):
		return reflect.TypeFor[float64]()
	case strings.Contains(dbType, "char"), strings.Contains(dbType, "text"), strings.Contains(dbType, "citext"):
		return reflect.TypeOf("")
	case strings.Contains(dbType, "time"), strings.Contains(dbType, "date"):
		return reflect.TypeOf(time.Time{})
	case strings.Contains(dbType, "json"), strings.Contains(dbType, "bytea"):
		return reflect.TypeOf([]byte{})
	default:
		// fallback
		return reflect.TypeOf(any(nil))
	}
}

// ---- convert "user_name" → "UserName" ----
func (h *helperType) toExportedName(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == ' ' || r == '-'
	})
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}
	return strings.Join(parts, "")
}

// ---- create struct ----
func (h *helperType) CreateDynamicStructFromSqlColumnType(colTypes []*sql.ColumnType) reflect.Type {
	fields := make([]reflect.StructField, 0, len(colTypes))

	for _, col := range colTypes {
		dbType := strings.ToUpper(col.DatabaseTypeName())
		goType := reflect.TypeOf(new(interface{})).Elem() // fallback

		switch dbType {
		case "INT", "INTEGER", "SMALLINT", "SERIAL", "BIGSERIAL":
			goType = reflect.TypeFor[int64]()
		case "BIGINT":
			goType = reflect.TypeFor[int64]()
		case "REAL", "FLOAT4":
			goType = reflect.TypeOf(float32(0))
		case "DOUBLE", "FLOAT8", "NUMERIC", "DECIMAL":
			goType = reflect.TypeFor[float64]()
		case "BOOLEAN", "BOOL":
			goType = reflect.TypeOf(false)
		case "CHAR", "VARCHAR", "TEXT", "CITEXT", "UUID":
			goType = reflect.TypeOf("")
		case "DATE", "TIMESTAMP", "TIMESTAMPTZ":
			goType = reflect.TypeOf(time.Time{})
		case "BYTEA":
			goType = reflect.TypeOf([]byte(nil))
		default:
			goType = reflect.TypeOf(new(interface{})).Elem()
		}

		field := reflect.StructField{
			Name: strings.Title(col.Name()), // viết hoa để export
			Type: goType,
			Tag:  reflect.StructTag(fmt.Sprintf(`db:"%s"`, col.Name())),
		}
		fields = append(fields, field)
	}

	return reflect.StructOf(fields)
}
func (h *helperType) createTypesInRowFromSqlColumnTypeInternal(key string, colTypes []*sql.ColumnType) []reflect.Type {
	ret := make([]reflect.Type, len(colTypes))
	for i, col := range colTypes {
		dbType := strings.ToUpper(col.DatabaseTypeName())
		goType := reflect.TypeOf(new(interface{})).Elem() // fallback

		switch dbType {
		case "INT", "INTEGER", "SMALLINT", "SERIAL", "BIGSERIAL":
			goType = reflect.TypeFor[int64]()
		case "BIGINT":
			goType = reflect.TypeFor[int64]()
		case "REAL", "FLOAT4":
			goType = reflect.TypeOf(float32(0))
		case "DOUBLE", "FLOAT8", "NUMERIC", "DECIMAL":
			goType = reflect.TypeFor[float64]()
		case "BOOLEAN", "BOOL":
			goType = reflect.TypeOf(false)
		case "CHAR", "VARCHAR", "TEXT", "CITEXT", "UUID":
			goType = reflect.TypeOf("")
		case "DATE", "TIMESTAMP", "TIMESTAMPTZ":
			goType = reflect.TypeOf(time.Time{})
		case "BYTEA":
			goType = reflect.TypeOf([]byte(nil))
		default:
			goType = reflect.TypeOf(new(interface{})).Elem()
		}
		ret[i] = goType

	}

	return ret
}

type initCreateTypesInRowFromSqlColumnType struct {
	val  []reflect.Type
	once sync.Once
}

var initCreateTypesInRowFromSqlColumnTypeCache sync.Map

func (h *helperType) CreateTypesInRowFromSqlColumnType(key string, colTypes []*sql.ColumnType) []reflect.Type {
	a, _ := initCreateTypesInRowFromSqlColumnTypeCache.LoadOrStore(key, &initCreateTypesInRowFromSqlColumnType{})
	i := a.(*initCreateTypesInRowFromSqlColumnType)
	i.once.Do(func() {
		i.val = h.createTypesInRowFromSqlColumnTypeInternal(key, colTypes)
	})
	return i.val
}
func (h *helperType) CreateRowsFromSqlColumnType(key string, colTypes []*sql.ColumnType) []reflect.Value {
	retTypes := h.CreateTypesInRowFromSqlColumnType(key, colTypes)
	ret := make([]reflect.Value, len(colTypes))
	//fields := make([]reflect.StructField, 0, len(colTypes))

	for i, goType := range retTypes {

		ret[i] = reflect.New(goType)

	}

	return ret
}
