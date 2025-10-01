package compiler

import (
	"fmt"
	"sync"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type cmpSelectorType struct {
	cmpType COMPILER
}

var CompilerSelect = &cmpSelectorType{}

type initMakeSelect struct {
	val  string
	err  error
	once sync.Once
}

var initMakeSelectCache sync.Map

func (cmp *cmpSelectorType) MakeSelect(dialect types.Dialect, outputFields *map[string]string, selectors, prefixKey string) (string, error) {
	key := selectors + "://" + prefixKey
	a, _ := initMakeSelectCache.LoadOrStore(key, &initMakeSelect{})
	i := a.(*initMakeSelect)
	i.once.Do(func() {
		i.val, i.err = cmp.makeSelectInternal(dialect, outputFields, selectors)
	})
	if i.err != nil {
		initMakeSelectCache.Delete(key)
	}
	return i.val, i.err
}
func (cmp *cmpSelectorType) makeSelectInternal(dialect types.Dialect, outputFields *map[string]string, selectors string) (string, error) {
	sql := "select " + selectors + " from tmp"
	sqlParse, err := internal.Helper.QuoteExpression(sql)
	if err != nil {
		return "", newCompilerError(fmt.Sprintf("'%s' is invalid syntax", selectors), ERR)
	}
	sqlExpr, err := sqlparser.Parse(sqlParse)
	if err != nil {
		return "", newCompilerError(fmt.Sprintf("'%s' is invalid syntax. Error:%s", selectors, err.Error()), ERR)
	}
	//*sqlparser.Select
	if selectExpr, ok := sqlExpr.(*sqlparser.Select); ok {

		ret, err := cmp.resolevSelector(dialect, outputFields, selectExpr.SelectExprs, selectors)
		return ret, err
	} else {
		return "", NewCompilerError(fmt.Sprintf("'%s' is invalid syntax", selectors))
	}

}
