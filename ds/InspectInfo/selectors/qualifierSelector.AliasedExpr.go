package selectors

import (
	"fmt"

	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/sqlparser"
)

// qualifierSelector.AliasedExpr.go
func (s *qualifierSelector) AliasedExpr(node *sqlparser.AliasedExpr, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	r, err := s.Expr(node.Expr, injectInfo)
	if err != nil {
		return nil, err
	}
	var alias string
	if node.As.IsEmpty() {
		alias = r.AliasField
		//r.Content = node.As.String()
	} else {
		alias = node.As.String()
	}
	r.Content = fmt.Sprintf("%s %s", r.Content, injectInfo.Dialect.Quote(alias))
	injectInfo.OuputFieldsInSelector = append(injectInfo.OuputFieldsInSelector, alias)

	return r, nil

}
