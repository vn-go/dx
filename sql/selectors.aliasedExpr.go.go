package sql

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

// selectors.aliasedExpr.go
func (s selectors) aliasedExpr(expr *sqlparser.AliasedExpr, injector *injector) (*compilerResult, error) {
	switch t := expr.Expr.(type) {
	case *sqlparser.ColName:
		return s.colName(t, injector)
	default:
		panic(fmt.Sprintf("unimplemented: %T. See selectors.aliasedExpr, %s", t, `sql\selectors.aliasedExpr.go.go`))

	}

}


