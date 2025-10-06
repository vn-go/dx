package compiler

import (
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) joinCondition(node sqlparser.JoinCondition, cmpType COMPILER, args *internal.SqlArgs) (string, error) {
	return cmp.resolve(node.On, cmpType, args)
}
