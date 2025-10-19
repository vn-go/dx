package quicky

import (
	"errors"
	"strings"

	"github.com/vn-go/dx/internal"
)

func (c *clause) getContentIn2Brackets(input string) (string, string) {
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

func (c *clause) Inspect(originalSql string) (*QueryItem, []string, error) {
	var err error
	inputSql := c.replaceQuestionMarks(originalSql, GET_PARAMS_FUNC)
	input, textParams := internal.Helper.InspectStringParam(inputSql)
	retPtr := c.ParseQueryItems(input)
	if retPtr == nil {
		return nil, nil, errors.New("ParseQueryItems failed")
	}

	ret := retPtr

	for retPtr != nil {

		retPtr.InspectData, err = c.ToSqlNode(retPtr.Content)
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
