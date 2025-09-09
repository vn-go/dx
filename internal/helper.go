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
	cache           sync.Map // map[string]string
	reSingleQuote   *regexp.Regexp
	reFieldAccess   *regexp.Regexp
	reFuncCall      *regexp.Regexp
	reFromJoin      *regexp.Regexp
	reAsAlias       *regexp.Regexp
	bufPool         *sync.Pool
}

// if s is "true" or "false" retun true
func (m *helperType) IsBool(s string) bool {
	return strings.ToLower(s) == "true" || strings.ToLower(s) == "false"
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
func (c *helperType) QuoteExpression(expr string) string {
	// Check cache
	if cached, ok := c.cache.Load(expr); ok {
		return cached.(string)
	}

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
	buf.WriteString(exprNoStr[lastPos:])

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
	c.cache.Store(expr, out)
	return out
}
func (c *helperType) AddrssertSinglePointerToStruct(obj interface{}) error {
	v := reflect.ValueOf(obj)
	t := v.Type()
	key := t.String() + "://helperType/AddrssertSinglePointerToStruct"
	_, err := OnceCall(key, func() (int, error) {
		depth := 0
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
			depth++
			if depth > 1 {
				break
			}
		}

		if depth != 1 {
			return depth, fmt.Errorf("Insert: expected pointer to struct (*T), got %d-level pointer", depth)
		}

		if t.Kind() != reflect.Struct {
			return depth, fmt.Errorf("Insert: expected pointer to struct, got pointer to %s", t.Kind())
		}
		return depth, nil

	})
	if err != nil {
		return err
	}

	return nil
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
			"limit": true,
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
