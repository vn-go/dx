package compiler

import (
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) joinTableExpr(node *sqlparser.JoinTableExpr, cmpType COMPILER) (string, error) {
	strCon, err := cmp.resolve(node.Condition, cmpType)
	if err != nil {
		return "", err
	}
	strLeft, err := cmp.resolve(node.LeftExpr, cmpType)
	if err != nil {
		return "", err
	}
	strRight, err := cmp.resolve(node.RightExpr, cmpType)
	if err != nil {
		return "", err
	}
	//fmt.Println(node.Condition.)
	// strJoin, err := cmp.resolve(node.Join, cmpType)
	// if err != nil {
	// 	return "", err
	// }
	return strLeft + " " + strings.ToUpper(node.Join) + " " + strRight + " ON " + strCon, nil
}
