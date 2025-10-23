package sql

import (
	"sync"

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

func (c compiler) Resolve(dialect types.Dialect, query string, arg ...any) (*sqlParser, error) {
	a, _ := initCompilerResolveCache.LoadOrStore(query, &initCompilerResolve{})
	i := a.(*initCompilerResolve)
	i.once.Do(func() {
		i.val, i.err = c.ResolveNoCache(dialect, query)
	})
	if i.err != nil {
		return nil, i.err
	}

	return &sqlParser{
		Query: i.val.Content,
		Args:  i.val.Args.ToArray(arg),
	}, nil

}
func (c compiler) ResolveNoCache(dialect types.Dialect, query string) (*compilerResult, error) {
	var err error
	//var node sqlparser.SQLNode
	var sqlStm sqlparser.Statement

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
