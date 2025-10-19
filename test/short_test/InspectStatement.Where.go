package shorttest

import (
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (fx InspectStatement) ExtractWhere(exprFn *sqlparser.FuncExpr, args *internal.SqlArgs) (*ResolveInfo, error) {
	return fx.Resolve(exprFn.Exprs[0], C_TYPE_WHERE, args)
}
