package selectors

import (
	"fmt"

	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/sqlparser"
)

// selector.SelectExpr.go
func (s *selector) SelectExpr(expr sqlparser.SelectExpr, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	switch expr := expr.(type) {
	case *sqlparser.AliasedExpr:
		return s.ResolveExpr(expr.Expr, injectInfo)
	default:
		panic(fmt.Sprintf("unsupported SelectExpr type %T, see: selector.SelectExpr, '%s'", expr, `ds\InspectInfo\selectors\selector.SelectExpr.go`))
	}
}
