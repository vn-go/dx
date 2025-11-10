package internal

import (
	"reflect"
	"regexp"
	sortTexts "sort"
	"strconv"
	"strings"
	"sync"

	"github.com/vn-go/dx/sqlparser"
)

var placeholderRegexp = regexp.MustCompile(`\{(\d+)\}`)

type ExtractParamMatrixInfo struct {
	Sql       string
	Data      map[int]int // newIndex -> oldIndex
	NewSize   int
	HasChange bool
}
type initExtractParamMatrix struct {
	val  *ExtractParamMatrixInfo
	err  error
	once sync.Once
}

var initExtractParamMatrixCach sync.Map

func (c *helperType) SortStrings(items []string) []string {
	sorted := make([]string, len(items))
	copy(sorted, items)
	sortTexts.Strings(sorted)
	return sorted
}

// ExtractParamMatrix builds a mapping from new arg indexes to original arg indexes.
// Example: {1}, {2}, {3}, {2} => map[0:0, 1:1, 2:2, 3:1]
func (c *helperType) ExtractParamMatrix(sql string) (*ExtractParamMatrixInfo, error) {
	a, _ := initExtractParamMatrixCach.LoadOrStore(sql, &initExtractParamMatrix{})
	i := a.(*initExtractParamMatrix)
	i.once.Do(func() {
		i.val, i.err = c.extractParamMatrix(sql)
	})

	return i.val, i.err

}
func (c *helperType) extractParamMatrixNoREgex(sql string) (*ExtractParamMatrixInfo, error) {
	ret := &ExtractParamMatrixInfo{
		Data: make(map[int]int),
	}

	currentNewIndex := 0
	inSingle := false
	inDouble := false
	result := make([]byte, 0, len(sql)+8)

	for i := 0; i < len(sql); {
		ch := sql[i]

		// Track quote state
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			result = append(result, ch)
			i++
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			result = append(result, ch)
			i++
			continue
		}

		// Replace {n} only outside quotes
		if !inSingle && !inDouble && ch == '{' && i+2 < len(sql) && sql[i+1] >= '0' && sql[i+1] <= '9' {
			j := i + 1
			num := 0
			for j < len(sql) && sql[j] >= '0' && sql[j] <= '9' {
				num = num*10 + int(sql[j]-'0')
				j++
			}
			if j < len(sql) && sql[j] == '}' && num > 0 {
				ret.Data[currentNewIndex] = num - 1
				currentNewIndex++
				result = append(result, '?')
				ret.HasChange = true
				i = j + 1
				continue
			}
		}

		result = append(result, ch)
		i++
	}

	ret.Sql = string(result)
	ret.NewSize = currentNewIndex
	return ret, nil
}
func (c *helperType) extractParamMatrix(sql string) (*ExtractParamMatrixInfo, error) {
	ret := &ExtractParamMatrixInfo{
		Data: make(map[int]int),
	}

	currentNewIndex := 0
	inSingle := false
	inDouble := false
	result := make([]byte, 0, len(sql)+8)

	for i := 0; i < len(sql); {
		ch := sql[i]

		// Track quote state
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			result = append(result, ch)
			i++
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			result = append(result, ch)
			i++
			continue
		}

		// Replace {n} only outside quotes
		if !inSingle && !inDouble && ch == '{' && i+2 < len(sql) && sql[i+1] >= '0' && sql[i+1] <= '9' {
			match := placeholderRegexp.FindStringSubmatch(sql[i:])
			if len(match) > 0 {
				index, err := strconv.Atoi(match[1])
				if err == nil && index >= 1 {
					ret.Data[currentNewIndex] = index - 1
					currentNewIndex++
					result = append(result, '?')
					ret.HasChange = true
					i += len(match[0])
					continue
				}
			}
		}

		result = append(result, ch)
		i++
	}

	ret.Sql = string(result)
	ret.NewSize = currentNewIndex
	return ret, nil
}

func (c *helperType) CombineType(t1, t2 reflect.Type, operator string) reflect.Type {
	if strings.ToLower(operator) == "and" {
		return c.ToNullableType(reflect.TypeOf(true))
	}
	if strings.ToLower(operator) == "or" {
		return c.ToNullableType(reflect.TypeOf(true))
	}
	if strings.ToLower(operator) == "not" {
		return c.ToNullableType(reflect.TypeOf(true))
	}
	var ret reflect.Type
	if t1 == nil && t2 == nil {
		return ret
	}
	if t1 == nil {
		return c.ToNullableType(t2)
	}
	return c.ToNullableType(t1)
}
func (c *helperType) CombineToDbTypeType(t1, t2 reflect.Type, operator string) sqlparser.ValType {
	return c.GetSqlTypeFfromGoType(c.CombineType(t1, t2, operator))
}
func (c *helperType) ToNullableType(t reflect.Type) reflect.Type {
	if t == nil {
		return nil
	}
	// Nếu đã là pointer hoặc interface thì coi như nullable rồi
	if t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		return t
	}
	// Còn lại thì lấy pointer của type
	return reflect.PtrTo(t)
}

// ApplyMatrix builds a new args slice from the matrix mapping.
// Example: Data={0:0,1:1,2:2,3:1}, oldArgs=[" ",2,3]
//
//	=> newArgs=[" ",2,3,2]
func (c *helperType) ApplyMatrix(info *ExtractParamMatrixInfo, oldArgs []any) []any {
	if info == nil || !info.HasChange {
		return oldArgs
	}
	newArgs := make([]any, info.NewSize)
	for newIdx, oldIdx := range info.Data {
		if oldIdx >= 0 && oldIdx < len(oldArgs) {
			newArgs[newIdx] = oldArgs[oldIdx]
		}
	}
	return newArgs
}
func (c *helperType) FixParam(sql string, inputArgs []any) (sqlOutput string, outputArgs []any, err error) {
	info, err := c.ExtractParamMatrix(sql)
	if err != nil {
		return "", nil, err
	}

	sqlOutput = info.Sql

	// Nếu không có placeholder hoặc không cần thay đổi
	if info.HasChange {
		outputArgs = c.ApplyMatrix(info, inputArgs)

		return
	}
	outputArgs = inputArgs

	return
}
