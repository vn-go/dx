package sql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type compiler struct {
}

func (c compiler) Resolve(dialect types.Dialect, query string, arg ...any) (*compilerResult, error) {
	var err error
	//var node sqlparser.SQLNode
	var sqlStm sqlparser.Statement
	inputSql := internal.Helper.ReplaceQuestionMarks(query, GET_PARAMS_FUNC)
	queryCompiling, textParams := internal.Helper.InspectStringParam(inputSql)
	injector := newInjector(dialect, textParams, arg)
	queryCompiling, err = internal.Helper.QuoteExpression(queryCompiling)
	if err != nil {
		return nil, err
	}
	sqlStm, err = sqlparser.Parse(queryCompiling)
	if err != nil {
		return nil, err
	}
	switch expr := sqlStm.(type) {
	case *sqlparser.Select:
		return selector.selects(expr, injector)
	case *sqlparser.Union:
		return selector.union(expr, injector)

	default:
		panic(fmt.Sprintf("not support statement type: %T. see compiler.Resolve", sqlStm))

	}

}

var Compiler = compiler{}
