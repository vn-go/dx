package selectors

import (
	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/sqlparser"
)

// qualifierSelector.BinaryExpr.go
func (s *qualifierSelector) BinaryExpr(expr *sqlparser.BinaryExpr, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	left, err := s.Expr(expr.Left, injectInfo)
	if err != nil {
		return nil, err
	}
	right, err := s.Expr(expr.Right, injectInfo)
	if err != nil {
		return nil, err
	}
	return &common.ResolverContent{
		Content:         left.Content + " " + expr.Operator + " " + right.Content,
		OriginalContent: left.OriginalContent + " " + expr.Operator + " " + right.OriginalContent,
	}, nil
}
