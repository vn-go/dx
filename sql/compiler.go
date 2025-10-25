package sql

import (
	"strings"
	"sync"
	"unicode"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type compiler struct {
}

type initCompilerResolve struct {
	val  *compilerResult
	err  error
	once sync.Once
}

var initCompilerResolveCache sync.Map

func (c compiler) Resolve(dialect types.Dialect, query string, arg ...any) (*SmartSqlParser, error) {
	a, _ := initCompilerResolveCache.LoadOrStore(query, &initCompilerResolve{})
	i := a.(*initCompilerResolve)
	i.once.Do(func() {
		i.val, i.err = c.ResolveNoCache(dialect, query)
	})
	if i.err != nil {
		return nil, i.err
	}
	args, err := i.val.Args.ToArray(arg)
	if err != nil {
		return nil, err
	}
	return &SmartSqlParser{
		Query:       i.val.Content,
		Args:        args,
		ScopeAccess: i.val.Fields,
	}, nil

}

// StartWithSelectKeyword kiểm tra xem chuỗi có bắt đầu bằng từ khóa "select"
// (bỏ qua khoảng trắng, tab, xuống dòng đầu chuỗi) và không có ký tự khác trước đó.
func (c compiler) startWithSelectKeyword(s string) bool {
	runes := []rune(s)
	n := len(runes)

	// Bỏ qua khoảng trắng, tab, xuống dòng ở đầu
	i := 0
	for i < n && unicode.IsSpace(runes[i]) {
		i++
	}

	// Nếu sau khi bỏ qua khoảng trắng mà bắt đầu bằng "select" (không phân biệt hoa thường)
	if i+6 <= n && strings.EqualFold(string(runes[i:i+6]), "select") {
		return true
	}

	// Nếu có ký tự khác hoặc không có "select" hợp lệ
	return false
}

func (c compiler) ResolveNoCache(dialect types.Dialect, query string) (*compilerResult, error) {
	var err error
	//var node sqlparser.SQLNode
	var sqlStm sqlparser.Statement
	if !c.startWithSelectKeyword(query) {
		querySimple, err := smartier.simple(query)
		if err != nil {
			return nil, err
		}

		query = querySimple
	}
	inputSql := internal.Helper.ReplaceQuestionMarks(query, GET_PARAMS_FUNC)
	queryCompiling, textParams := internal.Helper.InspectStringParam(inputSql)
	injector := newInjector(dialect, textParams)
	queryCompiling, err = internal.Helper.QuoteExpression(queryCompiling)
	if err != nil {
		return nil, err
	}
	sqlStm, err = sqlparser.Parse(queryCompiling)
	if err != nil {
		return nil, err
	}
	return froms.selectStatement(sqlStm, injector)

}

var Compiler = compiler{}
