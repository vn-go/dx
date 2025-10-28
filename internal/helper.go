package internal

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/vn-go/dx/sqlparser"
)

type helperType struct {
	SkipDefaulValue string
	keywords        map[string]bool
	keywords2       map[string]bool
	funcWhitelist   map[string]bool
	funcWhitelist2  map[string]bool

	reSingleQuote          *regexp.Regexp
	reFieldAccess          *regexp.Regexp
	reFuncCall             *regexp.Regexp
	reFromJoin             *regexp.Regexp
	reAsAlias              *regexp.Regexp
	bufPool                *sync.Pool
	mapGoTypeToSqlNodeType map[reflect.Type]sqlparser.ValType
}

func (h *helperType) GetSqlTypeFfromGoType(fieldType reflect.Type) sqlparser.ValType {
	typ := fieldType
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if ret, ok := h.mapGoTypeToSqlNodeType[typ]; ok {
		return ret
	} else {
		return -1
	}
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
	if val, ok := defaultValueByFromDbTag[defaultValue]; ok {
		return val, nil
	} else if strings.Contains(defaultValue, "'") {
		return defaultValue, nil
	}
	if m.IsFloatNumber(defaultValue) {
		return defaultValue, nil

	} else if m.IsNumber(defaultValue) {
		return defaultValue, nil

	} else if m.IsBool(defaultValue) {
		return defaultValue, nil

	} else {
		return "", fmt.Errorf("not support default value from %s, review GetGetDefaultValueByFromDbTag() function in %s", defaultValue, reflect.TypeOf(m).Elem())
	}
}
func EscapeQuote(s string) string {
	// Bước 1: thay \'
	s = replaceEscapedQuote(s)
	// Bước 2: URL encode
	return url.QueryEscape(s)
}

func replaceEscapedQuote(s string) string {
	out := make([]rune, 0, len(s))
	skip := false
	for i, r := range s {
		if skip {
			skip = false
			continue
		}
		if r == '\\' && i+1 < len(s) && s[i+1] == '\'' {
			out = append(out, '\'')
			skip = true
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

const FnMarkSpecialTextArgs = "dx__system_get_params_info"

// InspectStringParam scans an SQL statement and extracts all string literals
// enclosed in single quotes ('...'). Each detected literal is replaced by a
// parameter placeholder `?`, and all extracted literal values are returned
// as a slice of strings.
//
// Example:
//
//	sql := "SELECT * FROM a='hello ''jony''' AND b='ok'"
//	query, params := InspectStringParam(sql)
//
//	// Result:
//	// query  -> "SELECT * FROM a=? AND b=?"
//	// params -> []string{"hello 'jony'", "ok"}
//
// Rules:
//  1. A literal string starts and ends with a single quote (').
//  2. Two consecutive single quotes (”) inside a literal represent an escaped quote (').
//  3. The function scans linearly from left to right (O(n) complexity).
//  4. Only literal strings are replaced — numbers, NULLs, or identifiers remain unchanged.
//
// Notes:
//   - This is not a full SQL parser, but it correctly handles common SQL literal patterns.
//   - Works reliably for most SQL dialects (MySQL, PostgreSQL, SQL Server, etc.).
func (c *helperType) InspectStringParam(s string) (string, []string) {
	ret, _ := OnceCall(helperTypeKey+"/InspectStringParam/"+s, func() (*inspectStringParamStruct, error) {
		ret := &inspectStringParamStruct{}
		ret.Sql, ret.Params = c.inspectStringParam(s)
		return ret, nil
	})
	return ret.Sql, ret.Params

}

var helperTypeKey = reflect.TypeFor[helperType]().PkgPath() + "/" + reflect.TypeFor[helperType]().String()

type inspectStringParamStruct struct {
	Sql    string
	Params []string
}

func (c *helperType) inspectStringParam(s string) (string, []string) {
	var out strings.Builder // Holds the transformed SQL with ? placeholders
	var params []string     // Stores extracted string literal values
	inString := false       // Tracks whether the parser is currently inside a string literal
	var buf strings.Builder // Temporary buffer for building the current string literal
	index := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]

		if ch == '\'' { // Found a single quote
			if !inString {
				// Entering a string literal
				inString = true
				buf.Reset()
			} else {
				// Already inside a literal
				if i+1 < len(s) && s[i+1] == '\'' {
					// Found an escaped single quote ('')
					buf.WriteByte('\'')
					i++ // Skip the next quote
				} else {
					// End of the current literal
					inString = false
					params = append(params, buf.String())
					fnx := fmt.Sprintf("%s(%d)", FnMarkSpecialTextArgs, index)
					out.WriteString(fnx) // Replace literal with placeholder
					index++
				}
			}
			continue
		}

		if inString {
			// Collect characters inside the string literal
			buf.WriteByte(ch)
		} else {
			// Copy non-literal characters to output as-is
			out.WriteByte(ch)
		}
	}

	return out.String(), params
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

			// write content before token
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
		//out = strings.ReplaceAll(out, "\\'", "%27")
		return out, nil
	})

}
func (c *helperType) QuoteExpression2(expr string) (string, error) {

	return OnceCall("helperType/QuoteExpression2/"+expr, func() (string, error) {
		expr = strings.ReplaceAll(expr, "\n", " ")
		expr = strings.ReplaceAll(expr, "\t", " ")
		expr = strings.TrimSpace(expr)
		expr = strings.TrimSuffix(expr, ",")
		for strings.Contains(expr, ", ") {
			expr = strings.ReplaceAll(expr, ", ", ",")
		}
		for strings.Contains(expr, "  ") {
			expr = strings.ReplaceAll(expr, "  ", " ")
		}
		for strings.Contains(expr, "( ,") {
			expr = strings.ReplaceAll(expr, "( ,", "(")
		}
		for strings.Contains(expr, ", )") {
			expr = strings.ReplaceAll(expr, ", )", ")")
		}
		for strings.Contains(expr, "(,") {
			expr = strings.ReplaceAll(expr, "(,", "(")
		}
		for strings.Contains(expr, ",)") {
			expr = strings.ReplaceAll(expr, ",)", ")")
		}
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
			if c.keywords2[lowered] {
				continue
			}
			if c.funcWhitelist2[strings.ToLower(strings.Split(token, ".")[0])] {
				continue
			}

			// write content before token
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
		//out = strings.ReplaceAll(out, "\\'", "%27")
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

// ReplaceQuestionMarks thay thế các dấu '?' nằm ngoài chuỗi literal
// bằng functionName(pos), ví dụ: myFunc(1), myFunc(2), ...
func (c *helperType) ReplaceQuestionMarks(query, functionName string) string {
	var sb strings.Builder
	inString := false
	paramIndex := 1

	for i := 0; i < len(query); i++ {
		ch := query[i]

		// Toggle trạng thái trong chuỗi khi gặp dấu nháy đơn
		if ch == '\'' {
			sb.WriteByte(ch)
			// Xử lý trường hợp escape bằng 2 dấu nháy liên tiếp ('')
			if i+1 < len(query) && query[i+1] == '\'' {
				sb.WriteByte('\'')
				i++ // bỏ qua ký tự tiếp theo
			} else {
				inString = !inString
			}
			continue
		}

		// Nếu là dấu ? và không nằm trong chuỗi literal
		if ch == '?' && !inString {
			sb.WriteString(fmt.Sprintf("%s(%d)", functionName, paramIndex))
			paramIndex++
			continue
		}

		// Mặc định: ghi lại ký tự
		sb.WriteByte(ch)
	}

	return sb.String()
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
			FnMarkSpecialTextArgs: true,
		},
		keywords2: map[string]bool{
			"as": true, "and": true, "or": true, "not": true,
			"case": true, "when": true, "then": true, "else": true, "end": true,
			"inner": true, "left": true, "right": true, "full": true,
			"on": true, "using": true,
			"like":  true,
			"limit": true, "having": true, "is": true, "null": true, "offset": true,
			"delete": true, "update": true, "set": true,
			"range":               true,
			"to":                  true,
			"in":                  true,
			FnMarkSpecialTextArgs: true,
		},
		funcWhitelist: map[string]bool{
			"min": true, "max": true, "abs": true, "len": true,
			"sum": true, "avg": true, "count": true, "coalesce": true,
			"lower": true, "upper": true, "trim": true, "ltrim": true, "rtrim": true,
			"date_format": true, "date_add": true, "date_sub": true, "date": true,
			"year": true, "month": true, "day": true, "hour": true, "minute": true,
		}, funcWhitelist2: map[string]bool{
			// "min": true, "max": true, "abs": true, "len": true,
			// "sum": true, "avg": true, "count": true, "coalesce": true,
			// "lower": true, "upper": true, "trim": true, "ltrim": true, "rtrim": true,
			// "date_format": true, "date_add": true, "date_sub": true, "date": true,
			// "year": true, "month": true, "day": true, "hour": true, "minute": true,
		},
	}
	for k, v := range ret.funcWhitelist {
		ret.funcWhitelist[strings.ToUpper(k)] = v
	}
	for k, v := range ret.keywords {
		ret.keywords[strings.ToUpper(k)] = v
	}

	ret.reSingleQuote = regexp.MustCompile(`'(.*?)'`)
	ret.reFieldAccess = regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_.$]*\b`) //regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_\.]*\b`)
	ret.reFuncCall = regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\s*\(`)
	ret.reFromJoin = regexp.MustCompile(`(?i)(FROM|JOIN)\s+([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	ret.reAsAlias = regexp.MustCompile(`(?i)\bAS\s+([a-zA-Z_][a-zA-Z0-9_]*)\b`)
	ret.bufPool = &sync.Pool{New: func() any { return new(bytes.Buffer) }}
	ret.mapGoTypeToSqlNodeType = map[reflect.Type]sqlparser.ValType{
		reflect.TypeOf(int(0)):      sqlparser.IntVal,
		reflect.TypeOf(int8(0)):     sqlparser.IntVal,
		reflect.TypeOf(int16(0)):    sqlparser.IntVal,
		reflect.TypeOf(int32(0)):    sqlparser.IntVal,
		reflect.TypeOf(int64(0)):    sqlparser.IntVal,
		reflect.TypeOf(uint(0)):     sqlparser.IntVal,
		reflect.TypeOf(uint8(0)):    sqlparser.IntVal,
		reflect.TypeOf(uint16(0)):   sqlparser.IntVal,
		reflect.TypeOf(uint32(0)):   sqlparser.IntVal,
		reflect.TypeOf(uint64(0)):   sqlparser.IntVal,
		reflect.TypeOf(float32(0)):  sqlparser.FloatVal,
		reflect.TypeOf(float64(0)):  sqlparser.FloatVal,
		reflect.TypeOf(bool(false)): sqlparser.BitVal,
		reflect.TypeOf(string("")):  sqlparser.StrVal,
	}

	return ret
}
