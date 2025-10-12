package internal

import (
	"fmt"
	"reflect"
	"regexp"
)

// map kiểu Go -> kiểu PostgreSQL
func (h *helperType) postgresTypeFromGo(v any) string {
	if v == nil {
		return "text"
	}

	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "numeric"
	case reflect.Bool:
		return "bool"
	case reflect.String:
		return "citext"
	default:
		// Nếu kiểu là struct hoặc slice thì coi như text (tùy ORM bạn sẽ custom thêm)
		return "text"
	}
}

// thêm ::<type> vào từng $n tương ứng trong SQL
func (h *helperType) FixPostgresParamType(sql string, args []any) string {
	re := regexp.MustCompile(`\$(\d+)`)
	return re.ReplaceAllStringFunc(sql, func(m string) string {
		// lấy index param (1-based)
		var idx int
		fmt.Sscanf(m, "$%d", &idx)
		if idx <= len(args) {
			typ := h.postgresTypeFromGo(args[idx-1])
			return fmt.Sprintf("$%d::%s", idx, typ)
		}
		return m
	})
}
