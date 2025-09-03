package internal

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type exprUtils struct {
	keywords      map[string]bool
	funcWhitelist map[string]bool
	cache         sync.Map // map[string]string
	reSingleQuote *regexp.Regexp
	reFieldAccess *regexp.Regexp
	reFuncCall    *regexp.Regexp
	reFromJoin    *regexp.Regexp
	reAsAlias     *regexp.Regexp
	bufPool       *sync.Pool
}

func newExprUtils() *exprUtils {
	ret := &exprUtils{
		keywords: map[string]bool{
			"as": true, "and": true, "or": true, "not": true,
			"case": true, "when": true, "then": true, "else": true, "end": true,
			"inner": true, "left": true, "right": true, "full": true, "join": true,
			"on": true, "using": true, "where": true, "group": true, "by": true,
			"like": true,
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

// Replace quoted strings with <0>, <1>, etc., and return list of values
func (c *exprUtils) extractLiterals(expr string) (string, []string) {
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
func (c *exprUtils) QuoteExpression(expr string) string {
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

func (c *exprUtils) QuoteExpressionOld(expr string) string {
	// Check cache
	if cached, ok := c.cache.Load(expr); ok {
		return cached.(string)
	}
	expr = strings.ReplaceAll(expr, "\n", "")
	expr = strings.ReplaceAll(expr, "\t", "")
	expr = strings.TrimPrefix(expr, " ")
	expr = strings.TrimSuffix(expr, " ")
	expr = strings.TrimSuffix(expr, ",")
	// Get buffer from pool
	buf := c.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer c.bufPool.Put(buf)

	exprNoStr, literals := c.extractLiterals(expr)

	tokens := c.reFieldAccess.FindAllString(exprNoStr, -1)
	marked := make(map[string]bool)

	for _, token := range tokens {
		lowered := strings.ToLower(token)
		if c.keywords[lowered] || marked[token] {
			continue
		}

		if strings.Contains(exprNoStr, token+"(") {
			// likely a function
			fn := token
			if dot := strings.LastIndex(fn, "."); dot != -1 {
				fn = fn[dot+1:]
			}
			if c.funcWhitelist[fn] {
				marked[token] = true
				continue
			}
		}

		quoted := token
		if strings.Contains(token, ".") {
			parts := strings.Split(token, ".")
			for i := range parts {
				parts[i] = "`" + parts[i] + "`"
			}
			quoted = strings.Join(parts, ".")
		} else {
			quoted = "`" + token + "`"
		}

		exprNoStr = strings.ReplaceAll(exprNoStr, token, quoted)
		marked[token] = true
	}

	// Replace <i> placeholders with original literals
	buf.WriteString(exprNoStr)
	out := buf.String()
	for i, val := range literals {
		placeholder := "<" + strconv.Itoa(i) + ">" // Sửa lỗi ở đây
		out = strings.ReplaceAll(out, placeholder, "'"+val+"'")
	}
	out = strings.ReplaceAll(out, "[", "`")
	out = strings.ReplaceAll(out, "]", "`")
	// Save to cache
	c.cache.Store(expr, out)
	return out
}
