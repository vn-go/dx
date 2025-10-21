package helper

import (
	"fmt"
	"strings"
	"unicode"

	sysErrors "errors"

	"github.com/vn-go/dx/ds/errors"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type SqlNode struct {
	Nodes  []any
	source string
}
type MapSqlNode map[string]SqlNode
type InspectInfo struct {
	Texts []string
	Args  []any

	Content  string
	NextType string
	Next     *InspectInfo
	/*
		type of present is QueryItem or MapSqlNode
	*/
	InspectData MapSqlNode
}
type ContentInfo struct {
	Content         string
	OriginalContent string
}

const GET_PARAMS_FUNC = "dx__GetParams"

// ReplaceQuestionMarks thay thế các dấu '?' nằm ngoài chuỗi literal
// bằng functionName(pos), ví dụ: myFunc(1), myFunc(2), ...
func replaceQuestionMarks(query, functionName string) string {
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

func matchKeyword(runes []rune, i int, keyword string) bool {
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
func ParseQueryItems(input string) *InspectInfo {
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
				if matchKeyword(runes, i, "union") {
					// Kiểm tra có "union all" không
					j := i + len("union")
					for j < n && unicode.IsSpace(runes[j]) {
						j++
					}
					nextType := "union"
					if matchKeyword(runes, j, "all") {
						nextType = "union all"
						j += len("all")
					}

					left := strings.TrimSpace(string(runes[:i]))
					right := strings.TrimSpace(string(runes[j:]))

					return &InspectInfo{
						Content:  left,
						NextType: nextType,
						Next:     ParseQueryItems(right), // đệ quy parse phần sau
					}
				}
			}
		}
	}

	// Không có UNION nào ngoài ngoặc
	return &InspectInfo{
		Content: strings.TrimSpace(text),
	}
}

/*
examale:

	from(

			qr1 (
				from(incrementDetail)
				select(id,sum(amount) TotalAmount)
				where(id > 30)

			)

		)
	->map{
		"from": `qr1 (
				from(incrementDetail)
				select(id,sum(amount) TotalAmount)
				where(id > 30)

			)`
	}

	qr1 (
							from(incrementDetail)
							select(id,sum(amount) TotalAmount)
							where(id > 30)

						)
	->map{
		"qr1": `from(incrementDetail)
				select(id,sum(amount) TotalAmount)
				where(id > 30)`
*/
func toMap(input string) map[string]string {
	// fmt.Println("--------------------------------")
	// fmt.Println(input)
	// fmt.Println("-------------------------------")
	result := make(map[string]string)
	runes := []rune(input)
	n := len(runes)

	for i := 0; i < n; {
		// Bỏ qua khoảng trắng
		for i < n && unicode.IsSpace(runes[i]) {
			i++
		}
		if i >= n {
			break
		}

		// Đọc key (tên như qr1, from, select...)
		start := i
		for i < n && (unicode.IsLetter(runes[i]) || runes[i] == '_' || unicode.IsDigit(runes[i])) {
			i++
		}

		if i >= n {
			break
		}

		// 👇 Bỏ qua khoảng trắng giữa key và dấu '('
		for i < n && unicode.IsSpace(runes[i]) {
			i++
		}

		if i >= n || runes[i] != '(' {
			i++
			continue
		}

		key := strings.ToLower(string(runes[start:i]))
		i++ // bỏ qua '('

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

		if i >= n {
			break
		}

		value := strings.TrimSpace(string(runes[valueStart:i]))
		result[key] = value
		i++ // bỏ qua ')'
	}

	return result
}

func getContentIn2Brackets(input string) (string, string) {
	key := strings.Split(input, "(")[0]
	start := strings.Index(input, "(")
	if start == -1 {
		return key, ""
	}

	depth := 0
	for i := start; i < len(input); i++ {
		if input[i] == '(' {
			depth++
		} else if input[i] == ')' {
			depth--
			if depth == 0 {
				// Cắt phần nội dung bên trong
				return key, input[start+1 : i]
			}
		}
	}
	// Nếu không tìm thấy cặp ngoặc đóng hợp lệ
	return key, ""
}
func toSqlNode(input string) (MapSqlNode, error) {

	data := toMap(input)
	result := MapSqlNode{}
	var node any

	for key, value := range data {
		nodes := []any{}
		//expr := "select " + value
		expr, err := internal.Helper.QuoteExpression2(value)
		if err != nil {
			return nil, errors.NewParseError("clause %s: %s, error: %s", key, value, err)
		}
		node, err = sqlparser.Parse("select " + expr)
		if err != nil {
			guestKey, guestContent := getContentIn2Brackets(input)
			dataVal, _, errInspect := Inspect(guestContent)
			if errInspect != nil {
				return nil, errors.NewParseError("clause %s: %s, error: %s", key, value, err)
			}
			nodes = append(nodes, dataVal)
			result[strings.ToLower(strings.TrimSpace(guestKey))] = SqlNode{
				Nodes:  nodes,
				source: value,
			}

			//return result, nil
		} else {
			for _, x := range node.(*sqlparser.Select).SelectExprs {
				nodes = append(nodes, x)
			}
			result[strings.ToLower(strings.TrimSpace(key))] = SqlNode{
				Nodes:  nodes,
				source: value,
			}
		}

	}
	return result, nil
}
func Inspect(originalSql string) (*InspectInfo, []string, error) {

	var err error
	inputSql := replaceQuestionMarks(originalSql, GET_PARAMS_FUNC)
	input, textParams := internal.Helper.InspectStringParam(inputSql)
	retPtr := ParseQueryItems(input)
	if retPtr == nil {
		return nil, nil, sysErrors.New("ParseQueryItems failed")
	}

	ret := retPtr

	for retPtr != nil {

		retPtr.InspectData, err = toSqlNode(retPtr.Content)
		if err != nil {
			// try inspect retPtr.Content

			if err != nil {
				return nil, nil, err
			}

		}
		retPtr = retPtr.Next
	}

	return ret, textParams, nil
}
