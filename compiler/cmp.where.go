package compiler

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type inspectFilterResult struct {
	Fields []string
	Expr   *sqlparser.ComparisonExpr
}

// func (cmp *cmpWhereType) InspectFilter(filter string) (*inspectFilterResult, error) {
// 	var err error
// 	var sql = "select * from tmp where " + filter
// 	sql, err = internal.Helper.QuoteExpression(sql)
// 	if err != nil {
// 		return nil, newCompilerError(fmt.Sprintf("'%s' is invalid filter expession", filter), ERR)
// 	}

// 	stm, err := sqlparser.Parse(sql)
// 	if err != nil {
// 		return nil, newCompilerError(fmt.Sprintf("'%s' is invalid filter expession. Error:%s", filter, err.Error()), ERR)
// 	}
// 	expr := stm.(*sqlparser.Select).Where.Expr.(*sqlparser.ComparisonExpr)
// 	fields := cmp.getField(expr, make(map[string]bool))
// 	return &inspectFilterResult{
// 		Fields: fields,
// 		Expr:   expr,
// 	}, nil

// }
//
//	func (cmp *cmpWhereType) getField(n sqlparser.SQLNode, visited map[string]bool) []string {
//		if x, ok := n.(*sqlparser.ComparisonExpr); ok {
//			l := cmp.getField(x.Left, visited)
//			r := cmp.getField(x.Right, visited)
//			return append(l, r...)
//		}
//		if x, ok := n.(*sqlparser.ColName); ok {
//			if _, ok := visited[strings.ToLower(x.Name.String())]; !ok {
//				visited[strings.ToLower(x.Name.String())] = true
//				return []string{strings.ToLower(x.Name.String())}
//			}
//		}
//		if _, ok := n.(*sqlparser.SQLVal); ok {
//			return []string{}
//		}
//		if x, ok := n.(*sqlparser.BinaryExpr); ok {
//			l := cmp.getField(x.Left, visited)
//			r := cmp.getField(x.Right, visited)
//			return append(l, r...)
//		}
//		panic(fmt.Sprintf("Not implement,%T in %s", n, `compiler\cmp.where.go`))
//	}
type cmpWhereType struct {
}

var CmpWhere = &cmpWhereType{}

func (cmp *cmpWhereType) MakeFilter(dialect types.Dialect, outputFields map[string]string, filter string, numOfParams *int) (string, error) {
	sql := "select * from tmp where " + filter
	sqlParse, err := internal.Helper.QuoteExpression(sql)
	if err != nil {
		return "", newCompilerError(fmt.Sprintf("'%s' is invalid syntax", filter), ERR)
	}
	sqlExpr, err := sqlparser.Parse(sqlParse)
	if err != nil {
		return "", newCompilerError(fmt.Sprintf("'%s' is invalid syntax. Error:%s", filter, err.Error()), ERR)
	}
	//*sqlparser.Select

	return CompilerFilter.Resolve(dialect, filter, numOfParams, outputFields, sqlExpr.(*sqlparser.Select).Where.Expr)
}
