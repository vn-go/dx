package compiler

import "github.com/vn-go/dx/sqlparser"

func (cmp *compiler) comparisonExpr(node *sqlparser.ComparisonExpr, cmpType COMPILER) (string, error) {
	strLeft, err := cmp.resolve(node.Left, cmpType)
	if err != nil {
		return "", err
	}
	strRight, err := cmp.resolve(node.Right, cmpType)
	if err != nil {
		return "", err
	}
	return strLeft + " " + node.Operator + " " + strRight, nil
}
