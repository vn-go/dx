package compiler

import "github.com/vn-go/dx/sqlparser"

func (cmp *compiler) comparisonExpr(node *sqlparser.ComparisonExpr, cmpType COMPILER, args *[]any) (string, error) {
	strLeft, err := cmp.resolve(node.Left, cmpType, args)
	if err != nil {
		return "", err
	}
	strRight, err := cmp.resolve(node.Right, cmpType, args)
	if err != nil {
		return "", err
	}
	return strLeft + " " + node.Operator + " " + strRight, nil
}
