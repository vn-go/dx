package compiler

import "github.com/vn-go/dx/sqlparser"

func (cmp *compiler) andExpr(node *sqlparser.AndExpr, cmpType COMPILER, args *[]any) (string, error) {
	left, err := cmp.resolve(node.Left, cmpType, args)
	if err != nil {
		return "", err
	}
	right, err := cmp.resolve(node.Right, cmpType, args)
	if err != nil {
		return "", err
	}
	return left + " AND " + right, nil
}
func (cmp *compiler) orExpr(node *sqlparser.OrExpr, cmpType COMPILER, args *[]any) (string, error) {
	left, err := cmp.resolve(node.Left, cmpType, args)
	if err != nil {
		return "", err
	}
	right, err := cmp.resolve(node.Right, cmpType, args)
	if err != nil {
		return "", err
	}
	return left + " OR " + right, nil
}
func (cmp *compiler) notExpr(node *sqlparser.NotExpr, cmpType COMPILER, args *[]any) (string, error) {
	expr, err := cmp.resolve(node.Expr, cmpType, args)
	if err != nil {
		return "", err
	}

	return "NOT " + expr, nil
}
