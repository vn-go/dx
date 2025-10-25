package compiler

import (
	"fmt"
	"sync"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type paramInfo struct {
	Index int
}
type cmpSelectorType struct {
	cmpType       COMPILER
	aggregateExpr map[string]bool
	//args             internal.CompilerArgs
	originalSelector string
	//compilerStack    internal.Stack[COMPILER]
}

var CompilerSelect = &cmpSelectorType{}

type initMakeSelect struct {
	val  *ResolevSelectorResult
	err  error
	once sync.Once
}

var initMakeSelectCache sync.Map

func (cmp *cmpSelectorType) MakeSelect(dialect types.Dialect, outputFields *map[string]types.OutputExpr, selectors, prefixKey string, startOf2ApostropheArgs, startOfSqlIndex int) (*ResolevSelectorResult, error) {
	key := selectors + "://" + prefixKey
	a, _ := initMakeSelectCache.LoadOrStore(key, &initMakeSelect{})
	i := a.(*initMakeSelect)
	i.once.Do(func() {
		i.val, i.err = cmp.makeSelectInternal(dialect, outputFields, selectors, startOf2ApostropheArgs, startOfSqlIndex)

	})
	if i.err != nil {
		initMakeSelectCache.Delete(key)
		return nil, i.err
	}
	if i.val == nil {
		initMakeSelectCache.Delete(key)
		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", selectors))
	}
	return i.val, i.err
}

// type selectorResult struct {
// 	Expr  string
// 	GroupByFields map[string]string
// }

func (cmpTyp *cmpSelectorType) makeSelectInternal(dialect types.Dialect, outputFields *map[string]types.OutputExpr, selectors string, startOf2ApostropheArgs, startOfSqlIndex int) (*ResolevSelectorResult, error) {
	cmp := &cmpSelectorType{
		//compilerStack: internal.Stack[COMPILER]{},
	}
	cmp.originalSelector = selectors
	sqlPreprocess := "select " + selectors + " from t	mp"
	sql, apostropheArg := internal.Helper.InspectStringParam(sqlPreprocess)

	var args internal.SqlArgs = []internal.SqlArg{}

	sqlParse, err := internal.Helper.QuoteExpression(sql)

	if err != nil {

		return nil, newCompilerError(fmt.Sprintf("'%s' is invalid syntax", selectors), ERR)
	}

	sqlExpr, err := sqlparser.Parse(sqlParse)
	if err != nil {
		return nil, newCompilerError(fmt.Sprintf("'%s' is invalid syntax. Error:%s", selectors, err.Error()), ERR)
	}
	//*sqlparser.Select
	if selectExpr, ok := sqlExpr.(*sqlparser.Select); ok {
		ret, err := cmp.resolevSelector(dialect, outputFields, selectExpr.SelectExprs, selectors, &args, startOf2ApostropheArgs, startOfSqlIndex, apostropheArg)
		if err == nil {
			//no error get all args after compiler
			ret.Args = args
			ret.ApostropheArg = apostropheArg
			if ret.Selectors.HasAggregateFunction() {
				for _, x := range ret.Selectors {
					if x.FieldExprType == FieldExprType_Field {
						// if current selector is agg function call
						ret.GroupByExprs = append(ret.GroupByExprs, x.Expr)

					}
				}
				for _, v := range ret.FieldNotInAggFunc {
					ret.GroupByExprs = append(ret.GroupByExprs, v)
				}
			}

			return ret, nil
		} else {
			return nil, err
		}

	} else {
		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid syntax", selectors))
	}

}
