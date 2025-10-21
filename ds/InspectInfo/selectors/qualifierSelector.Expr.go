package selectors

import (
	"fmt"

	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/sqlparser"
)

// qualifierSelector.Expr.go
func (s *qualifierSelector) Expr(expr sqlparser.Expr, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	switch expr := expr.(type) {
	case *sqlparser.ColName:
		r, err := s.ColName(expr, injectInfo)
		if err != nil {
			return nil, err
		}

		return r, nil
	case *sqlparser.BinaryExpr:
		return s.BinaryExpr(expr, injectInfo)
	case *sqlparser.SQLVal:
		return common.Resolver.SQLVal(expr, injectInfo)
	default:
		panic(fmt.Sprintf("unimplemented node type: %T, see qualifierSelector.Expr, in file `%s`", expr, `ds\InspectInfo\selectors\qualifierSelector.Expr.go`))
	}

}
