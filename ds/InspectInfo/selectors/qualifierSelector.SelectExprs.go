package selectors

import (
	"fmt"

	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/sqlparser"
)

// qualifierSelector.ResolveSelectExprs.go
func (s *qualifierSelector) SelectExprs(exprs sqlparser.SelectExprs, injectInfo *common.InjectInfo) ([]string, error) {
	strColsSelected := []string{}
	for _, x := range exprs {
		switch x := x.(type) {
		case *sqlparser.AliasedExpr:
			r, err := s.AliasedExpr(x, injectInfo)
			if err != nil {
				return nil, err
			}
			strColsSelected = append(strColsSelected, r.Content)
		default:
			panic(fmt.Sprintf("unimplemented node type: %T, see qualifierSelector.ResolveSelectExprs, in file `%s`", x, `ds\InspectInfo\selectors\qualifierSelector.SelectExprs.go`))
		}

	}
	return strColsSelected, nil
}


