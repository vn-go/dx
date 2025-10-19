package quicky

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

func (s sqlNode) JoinClauseInFrom(dialect types.Dialect, textParams []string, dynamicArgs []any, arg *ArgInspects, field *FieldInspects, dict *Dictionanry) (string, error) {

	nodeOnClause := &JoinNode{}
	renderNodeOnClause := nodeOnClause
	var err error

	for _, node := range s.nodes {
		if aliasNode, ok := node.(*sqlparser.AliasedExpr); ok {
			if n, ok := aliasNode.Expr.(*sqlparser.ComparisonExpr); ok {
				renderNodeOnClause.Next = &JoinNode{
					Node:     n,
					JoinType: "INNER",
				}
				renderNodeOnClause = renderNodeOnClause.Next
				continue
			}
			if n, ok := aliasNode.Expr.(*sqlparser.ColName); ok {
				err = s.BuildDictionaryWithAlias(n, dialect, dict, aliasNode.As.String())
				if err != nil {
					return "", err
				}
				continue
			}
			//*sqlparser.ColName
			if n, ok := aliasNode.Expr.(*sqlparser.FuncExpr); ok {
				if n.Name.String() == "left" {
					renderNodeOnClause.Next = &JoinNode{
						Node:     n.Exprs[0],
						JoinType: "LEFT",
					}
					renderNodeOnClause = renderNodeOnClause.Next
				} else if n.Name.String() == "right" {
					renderNodeOnClause.Next = &JoinNode{
						Node:     n.Exprs[0],
						JoinType: "RIGHT",
					}
					renderNodeOnClause = renderNodeOnClause.Next
				} else if n.Name.String() == "full" {
					renderNodeOnClause.Next = &JoinNode{
						Node:     n.Exprs[0],
						JoinType: "FULL",
					}
					renderNodeOnClause = renderNodeOnClause.Next
				} else {
					err = s.BuildDictionary(aliasNode, dialect, dict)
					if err != nil {
						return "", err
					}

				}
			} else {
				err = s.BuildDictionary(aliasNode, dialect, dict)
				if err != nil {
					return "", err
				}

			}
		}

	}
	nodeOnClause = nodeOnClause.Next
	joinExprText := []string{}
	firstTable := ""
	previousTable := ""
	if nodeOnClause == nil {
		for _, x := range dict.AliasMap {
			joinExprText = append(joinExprText, dialect.Quote(x))

		}
		ret := strings.Join(joinExprText, "\n")
		return ret, nil
	} else {
		for nodeOnClause != nil {

			content, err := s.BuildOnJoin(nodeOnClause, dialect, textParams, dynamicArgs, arg, field, dict)
			if err != nil {
				return "", err
			}
			if firstTable == "" {
				firstTable = content.LeftTable
				txtJoin := fmt.Sprintf("%s\n %s JOIN %s ON %s", content.LeftTable, content.JoinType, content.RightTable, content.On)
				joinExprText = append(joinExprText, txtJoin)
				previousTable = content.RightTable
			} else {
				nextTable := content.RightTable
				if nextTable == previousTable {
					nextTable = content.LeftTable
				}
				txtJoin := fmt.Sprintf("%s JOIN %s ON %s", content.JoinType, nextTable, content.On)
				joinExprText = append(joinExprText, txtJoin)
			}

			nodeOnClause = nodeOnClause.Next

		}
		ret := strings.Join(joinExprText, "\n")

		return ret, err
	}
	return "", nil

}
