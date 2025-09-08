package compiler

import "github.com/vn-go/dx/sqlparser"

func (cmp *compiler) joinCondition(node sqlparser.JoinCondition, cmpType COMPILER) (string, error) {
	return cmp.resolve(node.On, cmpType)
}
