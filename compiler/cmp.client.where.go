package compiler

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type inspectFilterResult struct {
	Fields []string
	Expr   *sqlparser.ComparisonExpr
}

type cmpWhereType struct {
}

var CmpWhere = &cmpWhereType{}

type initMakeFilter struct {
	val  *CompilerFilterTypeResult
	err  error
	once sync.Once
}

var initMakeFilterCache sync.Map

func (cmp *cmpWhereType) MakeFilter(dialect types.Dialect, outputFields map[string]types.OutputExpr, filter string, sqlSource string, startOf2ApostropheArgs, startSqlIndex, startOdDynamicArg int) (*CompilerFilterTypeResult, error) {
	key := filter + "://" + reflect.TypeFor[cmpWhereType]().String() + "/" + sqlSource
	// for k, v := range outputFields {
	// 	key += k + "@" + v
	// }
	a, _ := initMakeFilterCache.LoadOrStore(key, &initMakeFilter{})
	i := a.(*initMakeFilter)
	i.once.Do(func() {
		i.val, i.err = cmp.makeFilterInternal(dialect, outputFields, filter, startOf2ApostropheArgs, startSqlIndex, startOdDynamicArg)
	})
	if i.err != nil {
		initMakeFilterCache.Delete(key)
		return nil, i.err
	}
	if i.val == nil {
		initMakeSelectCache.Delete(key)
		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", filter))
	}
	return i.val, i.err
}
func (cmp *cmpWhereType) makeFilterInternal(dialect types.Dialect, outputFields map[string]types.OutputExpr, filter string, startOf2ApostropheArgs, startSqlIndex, startOdDynamicArg int) (*CompilerFilterTypeResult, error) {
	var args internal.SqlArgs = []internal.SqlArg{}
	sqlPreProcess := "select * from tmp where " + filter
	sql, apostropheText := internal.Helper.InspectStringParam(sqlPreProcess)
	sqlParse, err := internal.Helper.QuoteExpression(sql)
	if err != nil {
		return nil, newCompilerError(fmt.Sprintf("'%s' is invalid syntax", filter), ERR)
	}
	sqlExpr, err := sqlparser.Parse(sqlParse)
	if err != nil {
		return nil, newCompilerError(fmt.Sprintf("'%s' is invalid syntax. Error:%s", filter, err.Error()), ERR)
	}
	//*sqlparser.Select
	if selectExpr, ok := sqlExpr.(*sqlparser.Select); ok {

		ret, err := CompilerFilter.Resolve(dialect, filter, outputFields, selectExpr.Where.Expr, &args, startOf2ApostropheArgs, startSqlIndex, startOdDynamicArg)
		if err != nil {
			return nil, err
		}
		ret.ApostropheArg = apostropheText
		ret.Args = args
		return ret, nil
	} else {
		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid syntax", filter))
	}

}
