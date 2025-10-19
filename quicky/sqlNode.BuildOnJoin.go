package quicky

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

// sqlNode.BuildOnJoin
func (s sqlNode) BuildOnJoin(nodeOnClause *JoinNode, dialect types.Dialect, textParams []string, dynamicArgs []any, arg *ArgInspects, field *FieldInspects, dict *Dictionanry) (*BuildOnJoinResult, error) {
	leftTable, rightTable, err := s.BuildOnJoinGetTables(nodeOnClause.Node, dialect, dict)
	if err != nil {
		return nil, err
	}
	content, err := resolver.Resolve(nodeOnClause.Node, dialect, textParams, dynamicArgs, arg, field, dict, C_TYPE_JOIN)
	if err != nil {
		return nil, err
	}

	//ret := fmt.Sprintf("%s  %s JOIN %s ON %s", leftTable, nodeOnClause.JoinType, rightTable, content)
	return &BuildOnJoinResult{
		LeftTable:  leftTable,
		RightTable: rightTable,
		On:         content,
		JoinType:   nodeOnClause.JoinType,
	}, nil

}

func (s sqlNode) BuildOnJoinGetTables(node sqlparser.SQLNode, dialect types.Dialect, dict *Dictionanry) (string, string, error) {
	switch node := node.(type) {
	case *sqlparser.AliasedExpr:
		return s.BuildOnJoinGetTables(node.Expr, dialect, dict)
	case *sqlparser.ComparisonExpr:
		leftTable, err := s.BuildOnJoinGetTable(node.Left, dialect, dict)
		if err != nil {
			return "", "", err
		}
		rightTable, err := s.BuildOnJoinGetTable(node.Right, dialect, dict)
		if err != nil {
			return "", "", err
		}
		return leftTable, rightTable, nil

	}
	panic(fmt.Sprintf("unhandled node type %T. See sqlNode.BuildOnJoinGetTables '%s'", node, `quicky\sqlNode.BuildOnJoin.go`))
}

func (s sqlNode) BuildOnJoinGetTable(node sqlparser.Expr, dialect types.Dialect, dict *Dictionanry) (string, error) {
	switch node := node.(type) {
	case *sqlparser.ColName:
		if node.Qualifier.IsEmpty() {
			return "", newParseError("can not determine dataset for column '%s'", node.Name.String())
		}
		tableName := strings.ToLower(node.Qualifier.Name.String())
		realTableName, ok := dict.AliasMap[tableName]

		if !ok {
			if err := s.BuildDictionary(node, dialect, dict); err != nil {
				return "", err
			}
			realTableName, ok = dict.AliasMap[tableName]
			if !ok {
				return "", newParseError("can not determine dataset for column '%s.%s'", node.Qualifier.Name.String(), node.Name.String())
			}
		}
		if aliasTable, ok := dict.AliasMapReverse[realTableName]; ok {

			retTableName := fmt.Sprintf("%s %s", dialect.Quote(realTableName), dialect.Quote(aliasTable))
			return retTableName, nil

		} else {
			return dialect.Quote(realTableName), nil

		}

	}
	panic(fmt.Sprintf("unhandled node type %T. See sqlNode.BuildOnJoinGetTable '%s'", node, `quicky\sqlNode.BuildOnJoin.go`))
}
