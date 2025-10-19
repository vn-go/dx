package quicky

import (
	"fmt"
	"strings"
	"unicode"
)

// ReplaceQuestionMarks thay thế các dấu '?' nằm ngoài chuỗi literal
// bằng functionName(pos), ví dụ: myFunc(1), myFunc(2), ...
func (c *clause) replaceQuestionMarks(query, functionName string) string {
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

func (c *clause) matchKeyword(runes []rune, i int, keyword string) bool {
	n := len(runes)
	k := len(keyword)
	if i+k > n {
		return false
	}
	word := string(runes[i : i+k])
	if strings.ToLower(word) != keyword {
		return false
	}

	// kiểm tra ký tự trước và sau không phải là chữ hoặc số
	if i > 0 && (unicode.IsLetter(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
		return false
	}
	if i+k < n && (unicode.IsLetter(runes[i+k]) || unicode.IsDigit(runes[i+k])) {
		return false
	}
	return true
}

/*
Example:

	select(user(name, age), department(name), u.id=d.id) where(u.age > 30))
	union
	select(customer(name, age), department(name), u.id=d.id) where(u.age > 30))

Output:

	QueryItem {
		Content: "select(user(name, age), department(name), u.id=d.id) where(u.age > 30))",
		NextType: "union",
		Next: QueryItem {
			Content: "select(customer(name, age), department(name), u.id=d.id) where(u.age > 30))",
			NextType: "",
			Next: nil,
		}
	}
*/
func (c *clause) ParseQueryItems(input string) *QueryItem {
	runes := []rune(input)
	n := len(runes)
	depth := 0

	// Chuẩn hóa chuỗi (bỏ whitespace dư)
	text := strings.TrimSpace(input)

	for i := 0; i < n; i++ {
		switch runes[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			// chỉ xét khi depth == 0
			if depth == 0 {
				// Kiểm tra từ khóa "union" tại vị trí này
				if c.matchKeyword(runes, i, "union") {
					// Kiểm tra có "union all" không
					j := i + len("union")
					for j < n && unicode.IsSpace(runes[j]) {
						j++
					}
					nextType := "union"
					if c.matchKeyword(runes, j, "all") {
						nextType = "union all"
						j += len("all")
					}

					left := strings.TrimSpace(string(runes[:i]))
					right := strings.TrimSpace(string(runes[j:]))

					return &QueryItem{
						Content:  left,
						NextType: nextType,
						Next:     c.ParseQueryItems(right), // đệ quy parse phần sau
					}
				}
			}
		}
	}

	// Không có UNION nào ngoài ngoặc
	return &QueryItem{
		Content: strings.TrimSpace(text),
	}
}
