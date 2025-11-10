package sql

import (
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type simpleSqlWithArgs struct {
	Content     string
	IndexOfArgs int
}
type simpleSql struct {
	from    string
	where   string
	selects string
	sort    string
	groupBy string
	offset  simpleSqlWithArgs
	limit   simpleSqlWithArgs
}
type initSimpleCache struct {
	val        string
	reIndex    []int
	textParams []string
	err        error
	once       sync.Once
}

var initSimpleCacheMap sync.Map

// return sql, index of offset in dynamic args, index of limit in dynamic args, error
func (s *smarty) simpleCache(simpleQuery string) (string, []int, []string, error) {
	a, _ := initSimpleCacheMap.LoadOrStore(simpleQuery, &initSimpleCache{})
	cache := a.(*initSimpleCache)
	cache.once.Do(func() {
		cache.val, cache.reIndex, cache.textParams, cache.err = s.simple(simpleQuery)
	})
	return cache.val, cache.reIndex, cache.textParams, cache.err
}

// return sql, index of offset in dynamic args, index of limit in dynamic args, error
func (s *smarty) simple(simpleQuery string) (string, []int, []string, error) {
	simpleQuery = internal.Helper.RemoveLineComments(simpleQuery)
	qr := simpleQuery
	_, textParams := internal.Helper.InspectStringParam(qr)
	str, err := internal.Helper.QuoteExpression2(qr)
	if err != nil {
		return "", nil, nil, err
	}

	tk := sqlparser.NewStringTokenizer("select " + str)
	stm, err := sqlparser.ParseNext(tk)

	if err != nil {
		return "", nil, nil, err
	}
	selectStm := stm.(*sqlparser.Select)
	ret, reindex, err := s.compile(selectStm, map[string]subsetInfo{})
	if err != nil {
		return "", nil, nil, err
	}
	return ret, reindex, textParams, nil

}
func Compact(simpleQuery string) (string, []int, []string, error) {
	return smartier.simple(simpleQuery)
}

// return sql, index of offset in dynamic args, index of limit in dynamic args, error
func (s *smarty) compile(selectStm *sqlparser.Select, refSubsets map[string]subsetInfo) (string, []int, error) {
	ret := &simpleSql{}
	subSetInfoList, err := subsets.extractSubSetInfo(selectStm, refSubsets)
	if err != nil {
		return "", nil, err
	}

	fieldAliasMap := map[string]string{}
	ret.selects, err = smartier.selectors(selectStm, fieldAliasMap, subSetInfoList)
	if err != nil {
		return "", nil, err
	}

	ret.where = smartier.where(selectStm)
	ret.groupBy = smartier.groupBy(selectStm, fieldAliasMap)
	ret.sort = smartier.sort(selectStm, fieldAliasMap)
	limitAndOffsetInfo, err := smartier.limitAndOffset(selectStm)
	if err != nil {
		return "", nil, err
	}
	if limitAndOffsetInfo != nil {
		ret.limit = simpleSqlWithArgs{
			Content:     limitAndOffsetInfo.take,
			IndexOfArgs: limitAndOffsetInfo.indexOfArgTake,
		}
		ret.offset = simpleSqlWithArgs{
			Content:     limitAndOffsetInfo.skip,
			IndexOfArgs: limitAndOffsetInfo.indexOfArgSkip,
		}
	}

	unionSource, err := unions.extractUnionInfo(selectStm, subSetInfoList)
	if err != nil {
		return "", nil, err
	}
	if unionSource != "" {
		if ret.where == "" && ret.groupBy == "" && ret.sort == "" && ret.selects == "" {
			return unionSource, nil, nil
		} else {
			if ret.selects == "*" {
				return unionSource, nil, nil
			}
			ret.from = "(" + unionSource + ") T"
		}
	} else if strFrom := smartier.from(selectStm, subSetInfoList); strFrom != "" {
		ret.from = strFrom
	} else {
		panic("can not detect from clause")
	}
	sqlText, newIndex := ret.String()

	return sqlText, newIndex, nil
}

type limitAndOffsetInfo struct {
	skip, take                     string
	indexOfArgSkip, indexOfArgTake int
}

func (s *smarty) limitAndOffset(selectStm *sqlparser.Select) (*limitAndOffsetInfo, error) {
	ret := &limitAndOffsetInfo{}
	found := false
	// var err error
	for _, x := range selectStm.SelectExprs {
		if fn := detect[*sqlparser.FuncExpr](x); fn != nil {
			found = true
			if fn.Name.Lowered() == "take" {
				ret.take = smartier.ToText(fn.Exprs[0])
				// if fnVal := detect[*sqlparser.SQLVal](fn.Exprs[0]); fnVal != nil {
				// 	ret.indexOfArgTake, err = internal.Helper.ToIntFormBytes(fnVal.Val[2:])
				// 	if err != nil {
				// 		return nil, err
				// 	}
				// 	ret.indexOfArgTake = ret.indexOfArgTake - 1
				// }
			}
			if fn.Name.Lowered() == "skip" {
				found = true
				ret.skip = smartier.ToText(fn.Exprs[0])
				// if fnVal := detect[*sqlparser.SQLVal](fn.Exprs[0]); fnVal != nil {
				// 	ret.indexOfArgSkip, err = internal.Helper.ToIntFormBytes(fnVal.Val[2:])
				// 	if err != nil {
				// 		return nil, err
				// 	}
				// 	ret.indexOfArgSkip = ret.indexOfArgSkip - 1
				// }
			}
		}
	}
	if !found {
		return nil, nil
	}
	return ret, nil
}

/*
	 Exmaple:
	 	sql="SELECT user.id, user.username FROM user WHERE id >= :v2 ORDER BY id asc LIMIT :v1 OFFSET :v3"
	 	sqlRet, newIndex := s.replaceVParams(sql)
		sqlRet: "SELECT user.id, user.username FROM user WHERE id >=? ORDER BY id asc LIMIT? OFFSET?"
		newIndex: [2 1 3]
*/
func (s *simpleSql) replaceVParamsOld(sql string) (string, []int) {
	var (
		result   strings.Builder
		inString bool // đang trong '...'
		runes    = []rune(sql)
		n        = len(runes)
		newIndex = []int{}
	)

	for i := 0; i < n; i++ {
		ch := runes[i]

		// Toggle khi gặp dấu nháy đơn (') không bị escape
		if ch == '\'' {
			inString = !inString
			result.WriteRune(ch)
			continue
		}

		// Nếu không trong chuỗi và gặp :v
		if !inString && ch == ':' && i+2 <= n && (i+1 < n && runes[i+1] == 'v') {

			j := i + 2
			// đọc hết phần số hoặc chữ
			for j < n && (unicode.IsDigit(runes[j]) || unicode.IsLetter(runes[j])) {
				j++
			}
			newIndex = append(newIndex, i)
			result.WriteRune('?')
			i = j - 1 // skip phần đã đọc
			continue
		}

		result.WriteRune(ch)
	}

	return result.String(), newIndex
}
func (s *simpleSql) replaceVParams(sql string) (string, []int) {
	var (
		result   strings.Builder
		inString bool // đang trong '...'
		runes    = []rune(sql)
		n        = len(runes)
		newIndex []int
	)

	for i := 0; i < n; i++ {
		ch := runes[i]

		// Toggle nếu gặp dấu nháy đơn mà không bị escape
		if ch == '\'' {
			inString = !inString
			result.WriteRune(ch)
			continue
		}

		// Nếu không trong chuỗi và gặp :v
		if !inString && ch == ':' && i+2 < n && runes[i+1] == 'v' {
			j := i + 2
			numStart := j
			for j < n && unicode.IsDigit(runes[j]) {
				j++
			}
			if numStart == j { // không có số sau v
				result.WriteRune(ch)
				continue
			}

			paramNumStr := string(runes[numStart:j])
			paramNum, err := strconv.Atoi(paramNumStr)
			if err == nil {
				newIndex = append(newIndex, paramNum-1)
			}

			result.WriteRune('?')
			i = j - 1 // bỏ qua phần số
			continue
		}

		result.WriteRune(ch)
	}

	return result.String(), newIndex
}

func (sql *simpleSql) String() (string, []int) {
	query := "SELECT " + sql.selects

	if sql.from != "" {
		query += " FROM " + sql.from
	}
	if sql.where != "" {
		query += " WHERE " + sql.where
	}
	if sql.groupBy != "" {
		query += " GROUP BY " + sql.groupBy
	}
	if sql.sort != "" {
		query += " ORDER BY " + sql.sort
	}

	// ⚡ Thêm LIMIT & OFFSET theo chuẩn MySQL
	if sql.limit.Content != "" {
		query += " LIMIT " + sql.limit.Content
		if sql.offset.Content != "" {
			query += " OFFSET " + sql.offset.Content
		}
	} else if sql.offset.Content != "" {
		// MySQL cho phép OFFSET mà không có LIMIT (dù ít khi dùng)
		query += " LIMIT 0 OFFSET " + sql.offset.Content
		// (số này là max uint64, tương đương "không giới hạn")
	}
	//return query
	sqlRet, newIndex := sql.replaceVParams(query)

	return sqlRet, newIndex
}
