package quicky

import (
	"strings"
	"unicode"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

var Clause = &clause{}

// ParseClauses đọc chuỗi DSL và trả về danh sách Clause
func (c *clause) ToArray(input string) []clauseItem {
	var result []clauseItem
	runes := []rune(input)
	n := len(runes)

	for i := 0; i < n; {
		// Bỏ qua khoảng trắng
		for i < n && unicode.IsSpace(runes[i]) {
			i++
		}

		// Đọc key
		start := i
		for i < n && (unicode.IsLetter(runes[i]) || runes[i] == '_') {
			i++
		}
		if i == start || i >= n || runes[i] != '(' {
			i++
			continue
		}

		key := strings.ToLower(string(runes[start:i]))

		// Bỏ qua '('
		i++
		depth := 1
		valueStart := i
		for i < n && depth > 0 {
			switch runes[i] {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					break
				}
			}
			i++
		}

		content := strings.TrimSpace(string(runes[valueStart:i]))

		result = append(result, clauseItem{
			Key:     key,
			Content: content,
		})

		// Bỏ qua dấu ')'
		i++
	}

	return result
}

func (c *clause) ToMap(input string) map[string]string {
	result := make(map[string]string)
	runes := []rune(input)
	n := len(runes)

	for i := 0; i < n; {
		// Bỏ qua khoảng trắng
		for i < n && unicode.IsSpace(runes[i]) {
			i++
		}

		// Đọc key (tên clause)
		start := i
		for i < n && (unicode.IsLetter(runes[i]) || runes[i] == '_') {
			i++
		}
		if i == start || i >= n || runes[i] != '(' {
			i++
			continue
		}

		key := strings.ToLower(string(runes[start:i])) // ví dụ "select"

		// Bỏ qua dấu '('
		i++
		depth := 1
		valueStart := i
		for i < n && depth > 0 {
			if runes[i] == '(' {
				depth++
			} else if runes[i] == ')' {
				depth--
				if depth == 0 {
					break
				}
			}
			i++
		}
		value := strings.TrimSpace(string(runes[valueStart:i]))
		result[key] = value

		// Bỏ qua dấu ')'
		i++
	}

	return result
}
func (c *clause) ToSqlNode(input string) (MapSqlNode, error) {
	data := c.ToMap(input)
	result := MapSqlNode{}
	var node any

	for key, value := range data {
		nodes := []any{}
		//expr := "select " + value
		expr, err := internal.Helper.QuoteExpression2(value)
		if err != nil {
			return nil, newParseError("clause %s: %s, error: %s", key, value, err)
		}
		node, err = sqlparser.Parse("select " + expr)
		if err != nil {
			guestKey, guestContent := c.getContentIn2Brackets(input)
			dataVal, _, errInspect := c.Inspect(guestContent)
			if errInspect != nil {
				return nil, newParseError("clause %s: %s, error: %s", key, value, err)
			}
			nodes = append(nodes, dataVal)
			result[strings.ToLower(guestKey)] = sqlNode{
				nodes:  nodes,
				source: value,
			}

			//return result, nil
		} else {
			for _, x := range node.(*sqlparser.Select).SelectExprs {
				nodes = append(nodes, x)
			}
			result[strings.ToLower(key)] = sqlNode{
				nodes:  nodes,
				source: value,
			}
		}

	}
	return result, nil
}
