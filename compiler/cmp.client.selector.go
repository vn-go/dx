package compiler

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type cmpSelectorType struct {
}

func (cmp *cmpSelectorType) MakeSelect(dialect types.Dialect, outputFields map[string]string, selectors string, numOfParams *int) (string, error) {
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
		return cmp.resolevSelector(dialect, outputFields, selectExpr.SelectExprs)
	} else {
		return "", NewCompilerError(fmt.Sprintf("'%s' is invalid syntax", selectors))
	}

}
