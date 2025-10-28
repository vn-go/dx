package sql

import (
	"strings"
	"sync"
	"unicode"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type simpleSql struct {
	from    string
	where   string
	selects string
	sort    string
	groupBy string
	offset  string
	limit   string
}
type initSimpleCache struct {
	val  string
	err  error
	once sync.Once
}

var initSimpleCacheMap sync.Map

func (s *smarty) simpleCache(simpleQuery string) (string, error) {
	a, _ := initSimpleCacheMap.LoadOrStore(simpleQuery, &initSimpleCache{})
	cache := a.(*initSimpleCache)
	cache.once.Do(func() {
		cache.val, cache.err = s.simple(simpleQuery)
	})
	return cache.val, cache.err
}

func (s *smarty) simple(simpleQuery string) (string, error) {
	str, err := internal.Helper.QuoteExpression2(simpleQuery)
	if err != nil {
		return "", err
	}

	tk := sqlparser.NewStringTokenizer("select " + str)
	stm, err := sqlparser.ParseNext(tk)

	if err != nil {
		return "", err
	}
	selectStm := stm.(*sqlparser.Select)
	return s.compile(selectStm, map[string]subsetInfo{})

}

func (s *smarty) compile(selectStm *sqlparser.Select, refSubsets map[string]subsetInfo) (string, error) {
	ret := &simpleSql{}
	subSetInfoList, err := subsets.extractSubSetInfo(selectStm, refSubsets)
	if err != nil {
		return "", err
	}

	fieldAliasMap := map[string]string{}
	ret.selects, err = smartier.selectors(selectStm, fieldAliasMap)
	if err != nil {
		return "", err
	}

	ret.where = smartier.where(selectStm)
	ret.groupBy = smartier.groupBy(selectStm, fieldAliasMap)
	ret.sort = smartier.sort(selectStm, fieldAliasMap)
	ret.offset, ret.limit = smartier.limitAndOffset(selectStm)
	unionSource, err := unions.extractUnionInfo(selectStm, subSetInfoList)
	if err != nil {
		return "", err
	}
	if unionSource != "" {
		if ret.where == "" && ret.groupBy == "" && ret.sort == "" {
			return unionSource, nil
		} else {
			ret.from = "(" + unionSource + ")"
		}
	} else if strFrom := smartier.from(selectStm, subSetInfoList); strFrom != "" {
		ret.from = strFrom
	} else {
		panic("can not detect from clause")
	}
	sqlText := ret.String()

	return sqlText, nil
}

func (s *smarty) limitAndOffset(selectStm *sqlparser.Select) (skip string, take string) {

	for _, x := range selectStm.SelectExprs {
		if fn := detect[*sqlparser.FuncExpr](x); fn != nil {
			if fn.Name.Lowered() == "take" {
				take = smartier.ToText(fn.Exprs[0])
			}
			if fn.Name.Lowered() == "skip" {
				skip = smartier.ToText(fn.Exprs[0])
			}
		}
	}
	return
}

func (s *simpleSql) replaceVParams(sql string) string {
	var (
		result   strings.Builder
		inString bool // đang trong '...'
		runes    = []rune(sql)
		n        = len(runes)
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

			result.WriteRune('?')
			i = j - 1 // skip phần đã đọc
			continue
		}

		result.WriteRune(ch)
	}

	return result.String()
}
func (sql *simpleSql) String() string {
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
	if sql.limit != "" {
		query += " LIMIT " + sql.limit
		if sql.offset != "" {
			query += " OFFSET " + sql.offset
		}
	} else if sql.offset != "" {
		// MySQL cho phép OFFSET mà không có LIMIT (dù ít khi dùng)
		query += " LIMIT 0 OFFSET " + sql.offset
		// (số này là max uint64, tương đương "không giới hạn")
	}

	return sql.replaceVParams(query)
}
