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
	ret.from = smartier.from(selectStm)

	ret.selects = smartier.selectors(selectStm,fieldAliasMap)

	ret.where = smartier.where(selectStm)
	ret.groupBy = smartier.groupBy(selectStm, fieldAliasMap)
	ret.sort = smartier.sort(selectStm,fieldAliasMap)
	return ret.String(), nil
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

	return sql.replaceVParams(query)
}
