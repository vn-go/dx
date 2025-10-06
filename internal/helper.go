package internal

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type helperType struct {
	SkipDefaulValue string
	keywords        map[string]bool
	funcWhitelist   map[string]bool

	reSingleQuote *regexp.Regexp
	reFieldAccess *regexp.Regexp
	reFuncCall    *regexp.Regexp
	reFromJoin    *regexp.Regexp
	reAsAlias     *regexp.Regexp
	bufPool       *sync.Pool
}

// if s is "true" or "false" retun true
func (m *helperType) IsBool(s string) bool {
	return strings.ToLower(s) == "true" || strings.ToLower(s) == "false" || strings.ToLower(s) == "yes" || strings.ToLower(s) == "no"
}
func (m *helperType) IsString(s string) bool {
	return s[0] == '\'' && s[len(s)-1] == '\''
}
func (m *helperType) IsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
func (m *helperType) IsFloatNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// Replace quoted strings with <0>, <1>, etc., and return list of values
func (c *helperType) extractLiterals(expr string) (string, []string) {
	var params []string
	var buf bytes.Buffer
	pos := 0
	matches := c.reSingleQuote.FindAllStringSubmatchIndex(expr, -1)

	for _, m := range matches {
		start, end := m[0], m[1]
		buf.WriteString(expr[pos:start])
		buf.WriteByte('<')
		buf.WriteByte(byte('0' + len(params)))
		buf.WriteByte('>')
		params = append(params, expr[m[2]:m[3]])
		pos = end
	}
	buf.WriteString(expr[pos:])
	return buf.String(), params
}
func (m *helperType) GetDefaultValue(defaultValue string, defaultValueByFromDbTag map[string]string) (string, error) {
	if strings.Contains(defaultValue, "'") {
		return defaultValue, nil
	}
	if m.IsFloatNumber(defaultValue) {
		return defaultValue, nil

	} else if m.IsNumber(defaultValue) {
		return defaultValue, nil

	} else if m.IsBool(defaultValue) {
		return defaultValue, nil

	} else if val, ok := defaultValueByFromDbTag[defaultValue]; ok {
		return val, nil
	} else {
		return "", fmt.Errorf("not support default value from %s, review GetGetDefaultValueByFromDbTag() function in %s", defaultValue, reflect.TypeOf(m).Elem())
	}
}
func (c *helperType) QuoteExpression(expr string) (string, error) {

	return OnceCall("helperType/QuoteExpression/"+expr, func() (string, error) {
		expr = strings.ReplaceAll(expr, "\n", " ")
		expr = strings.ReplaceAll(expr, "\t", " ")
		expr = strings.TrimSpace(expr)
		expr = strings.TrimSuffix(expr, ",")

		exprNoStr, literals := c.extractLiterals(expr)

		// Lấy từng token và vị trí
		matches := c.reFieldAccess.FindAllStringIndex(exprNoStr, -1)

		// Sử dụng buffer để build lại chuỗi
		buf := c.bufPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer c.bufPool.Put(buf)

		lastPos := 0
		for _, match := range matches {
			start, end := match[0], match[1]
			token := exprNoStr[start:end]

			lowered := strings.ToLower(token)
			if c.keywords[lowered] {
				continue
			}
			if c.funcWhitelist[strings.ToLower(strings.Split(token, ".")[0])] {
				continue
			}

			// Ghi đoạn trước token
			buf.WriteString(exprNoStr[lastPos:start])

			// Quote token
			parts := strings.Split(token, ".")
			for i, p := range parts {
				if i > 0 {
					buf.WriteByte('.')
				}
				buf.WriteString("`")
				buf.WriteString(p)
				buf.WriteString("`")
			}

			lastPos = end
		}

		// Ghi phần còn lại
		_, err := buf.WriteString(exprNoStr[lastPos:])
		if err != nil {
			return "", err
		}

		// Khôi phục các chuỗi literal
		out := buf.String()
		for i, val := range literals {
			placeholder := "<" + strconv.Itoa(i) + ">"
			out = strings.ReplaceAll(out, placeholder, "'"+val+"'")
		}

		// Chuyển [] sang ``
		out = strings.ReplaceAll(out, "[", "`")
		out = strings.ReplaceAll(out, "]", "`")

		// Cache kết quả

		return out, nil
	})

}

type initAddrssertSinglePointerToStruct struct {
	err  error
	once sync.Once
}

var cacheAddrssertSinglePointerToStruct sync.Map

func (c *helperType) AddrssertSinglePointerToStruct(obj interface{}) error {
	v := reflect.ValueOf(obj)
	t := v.Type()
	key := t.String() + "://helperType/AddrssertSinglePointerToStruct"
	actually, _ := cacheAddrssertSinglePointerToStruct.LoadOrStore(key, &initAddrssertSinglePointerToStruct{})
	init := actually.(*initAddrssertSinglePointerToStruct)
	init.once.Do(func() {
		depth := 0
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
			depth++
			if depth > 1 {
				break
			}
		}

		if depth != 1 {
			init.err = fmt.Errorf("expected pointer to struct (*T), got %d-level pointer", depth)
		}

		if t.Kind() != reflect.Struct {
			init.err = fmt.Errorf("expected pointer to struct, got pointer to %s", t.Kind())
		}

	})

	return init.err
}

var cacheAddrssertSinglePointerToSlice sync.Map

func (c *helperType) AddrssertSinglePointerToSlice(obj interface{}) error {
	v := reflect.ValueOf(obj)
	t := v.Type()
	key := t.String() + "://helperType/AddrssertSinglePointerToSlice"
	actually, _ := cacheAddrssertSinglePointerToSlice.LoadOrStore(key, &initAddrssertSinglePointerToStruct{})
	init := actually.(*initAddrssertSinglePointerToStruct)
	init.once.Do(func() {

		depth := 0
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
			depth++
			if depth > 1 {
				break
			}
		}

		if depth != 1 {
			init.err = fmt.Errorf("expected pointer to slice (*T), got %d-level pointer", depth)
		}

		if t.Kind() != reflect.Slice {
			init.err = fmt.Errorf(" expected pointer to slice, got pointer to %s", t.Kind())
		}
		if t.Elem().Kind() != reflect.Struct {
			init.err = fmt.Errorf(" expected pointer to slice of struct , got pointer to %s", t.String())
		}

	})

	return init.err
}

type intHelperTypeFindField struct {
	fieldIndex []int
	fieldType  reflect.Type
	found      bool
	once       sync.Once
}

var cacheHelperTypeFindField sync.Map

func (c *helperType) FindField(typ reflect.Type, fieldName string) ([]int, reflect.Type, bool) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	actually, _ := cacheHelperTypeFindField.LoadOrStore(typ.String()+"//"+strings.ToLower(fieldName), &intHelperTypeFindField{})
	item := actually.(*intHelperTypeFindField)
	item.once.Do(func() {
		field, ok := typ.FieldByNameFunc(func(s string) bool {
			return unicode.IsUpper([]rune(s)[0]) && strings.EqualFold(s, fieldName)
		})
		if ok {
			item.fieldIndex = field.Index
			item.found = ok
			item.fieldType = field.Type
			if item.fieldType.Kind() == reflect.Ptr {
				item.fieldType = item.fieldType.Elem()
			}
		}
	})
	return item.fieldIndex, item.fieldType, item.found
}

// Hàm kiểm tra tên trường hợp lệ
func (c *helperType) IsValidFieldName(field string) bool {
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	return re.MatchString(field)
}
func (c *helperType) GetAlias(txt string) string {
	items := strings.Split(txt, " ")
	if len(items) < 2 {
		return ""
	}
	if c.IsValidFieldName(items[len(items)-1]) {
		return items[len(items)-1]
	}
	return ""
}
func (c *helperType) TrimStringLiteral(s string) string {
	if len(s) == 0 {
		return s
	}
	if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
		s = s[1 : len(s)-1]
	}
	// Chuẩn hóa escape ký tự nháy đơn trong SQL
	s = strings.ReplaceAll(s, "''", "'")
	s = strings.ReplaceAll(s, `\"`, `"`)
	return s
}
func (c *helperType) ToBoolFromBytes(bff []byte) bool {
	return c.ToBool(string(bff))
}
func (c *helperType) ToBool(v string) bool {
	if v == "" {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "t", "true", "y", "yes", "on":
		return true
	case "0", "f", "false", "n", "no", "off":
		return false
	default:
		// Nếu không nhận dạng được, có thể coi là false
		return false
	}
}
func (c *helperType) ToInt(strV string) (int, error) {
	// Convert the byte slice to string

	// Trim leading and trailing spaces (in case there are spaces or newlines)
	strV = strings.TrimSpace(strV)

	// Parse the string into an integer
	i, err := strconv.Atoi(strV)
	if err != nil {
		// Return 0 and the parsing error if conversion fails
		return 0, fmt.Errorf("failed to convert bytes to int: %w", err)
	}

	// Return the parsed integer value and nil error
	return i, nil
}
func (c *helperType) ToIntFormBytes(v []byte) (int, error) {
	// Convert the byte slice to string
	strV := string(v)

	// Trim leading and trailing spaces (in case there are spaces or newlines)
	strV = strings.TrimSpace(strV)

	// Parse the string into an integer
	i, err := strconv.Atoi(strV)
	if err != nil {
		// Return 0 and the parsing error if conversion fails
		return 0, fmt.Errorf("failed to convert bytes to int: %w", err)
	}

	// Return the parsed integer value and nil error
	return i, nil
}
func (c *helperType) ToFloat(strV string) (float64, error) {
	// Convert byte slice to string

	// Trim spaces and newlines to ensure clean input
	strV = strings.TrimSpace(strV)

	// Parse string to float64
	f, err := strconv.ParseFloat(strV, 64)
	if err != nil {
		// Return 0 and wrap the error for context
		return 0, fmt.Errorf("failed to convert bytes to float64: %w", err)
	}

	// Return the parsed float64 value and nil error
	return f, nil
}
func (c *helperType) ToFloatFormBytes(v []byte) (float64, error) {
	// Convert byte slice to string
	strV := string(v)

	// Trim spaces and newlines to ensure clean input
	strV = strings.TrimSpace(strV)

	// Parse string to float64
	f, err := strconv.ParseFloat(strV, 64)
	if err != nil {
		// Return 0 and wrap the error for context
		return 0, fmt.Errorf("failed to convert bytes to float64: %w", err)
	}

	// Return the parsed float64 value and nil error
	return f, nil
}

var Helper = newHelper()

func newHelper() *helperType {
	ret := &helperType{
		SkipDefaulValue: "vdb::skip",
		keywords: map[string]bool{
			"as": true, "and": true, "or": true, "not": true,
			"case": true, "when": true, "then": true, "else": true, "end": true,
			"inner": true, "left": true, "right": true, "full": true, "join": true,
			"on": true, "using": true, "where": true, "group": true, "by": true,
			"like": true, "desc": true, "asc": true, "select": true, "from": true, "order": true,
			"union": true, "all": true,
			"limit": true, "having": true, "is": true, "null": true, "offset": true,
			"delete": true, "update": true, "set": true,
		},
		funcWhitelist: map[string]bool{
			"min": true, "max": true, "abs": true, "len": true,
			"sum": true, "avg": true, "count": true, "coalesce": true,
			"lower": true, "upper": true, "trim": true, "ltrim": true, "rtrim": true,
			"date_format": true, "date_add": true, "date_sub": true, "date": true,
			"year": true, "month": true, "day": true, "hour": true, "minute": true,
		},
	}
	for k, v := range ret.funcWhitelist {
		ret.funcWhitelist[strings.ToUpper(k)] = v
	}
	for k, v := range ret.keywords {
		ret.keywords[strings.ToUpper(k)] = v
	}

	ret.reSingleQuote = regexp.MustCompile(`'(.*?)'`)
	ret.reFieldAccess = regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_\.]*\b`)
	ret.reFuncCall = regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\s*\(`)
	ret.reFromJoin = regexp.MustCompile(`(?i)(FROM|JOIN)\s+([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	ret.reAsAlias = regexp.MustCompile(`(?i)\bAS\s+([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	ret.bufPool = &sync.Pool{New: func() any { return new(bytes.Buffer) }}
	return ret
}
