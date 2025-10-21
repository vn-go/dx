package from

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/ds/errors"
	"github.com/vn-go/dx/ds/helper"
	"github.com/vn-go/dx/sqlparser"
)

type fromClauseType struct {
}

var FromClause = &fromClauseType{}

type ResolveResult struct {
	FromClause string
	Dict       common.Dictionary
}
type fromClauseInfo struct {
	left     string
	right    string
	on       string
	joinType string
	next     *fromClauseInfo
	top      *fromClauseInfo
}

func (f *fromClauseType) Resolve(nodes []any, injectInfo *common.InjectInfo) (
	string, error) {

	var err error
	var navigateFromClauseInfo *fromClauseInfo
	var fromClauseInfo *fromClauseInfo
	// injectInfo := &InjectInfo{
	// 	Args:        []common.ArgScaner{},
	// 	DyanmicArgs: args,
	// 	TextArgs:    texts,
	// 	Dialect:     dialect,
	// 	Scope:       scope,
	// 	Dict:        NewDictionary(),
	// }
	sourceTables := []string{}
	for _, node := range nodes {
		switch node := node.(type) {
		case *sqlparser.AliasedExpr:
			switch expr := node.Expr.(type) {
			case *sqlparser.ColName:
				injectInfo.Dict.BuildByAliasTableName(injectInfo.Dialect, node.As.String(), expr.Name.String())
				if entityName, ok := injectInfo.Dict.Entities[strings.ToLower(expr.Name.String())]; ok {
					tableName := entityName.TableName
					if alias, ok := injectInfo.Dict.TableAlias[strings.ToLower(tableName)]; ok {
						if alias != "" {
							sourceTables = append(sourceTables, injectInfo.Dialect.Quote(tableName)+" "+injectInfo.Dialect.Quote(alias))
						} else {
							sourceTables = append(sourceTables, injectInfo.Dialect.Quote(tableName))
						}
					} else {
						sourceTables = append(sourceTables, injectInfo.Dialect.Quote(tableName))
					}
				}

				// TODO: resolve colname
			case *sqlparser.FuncExpr:
				if navigateFromClauseInfo == nil {
					navigateFromClauseInfo, err = f.buildFromClauseInfo(expr, injectInfo)
					if err != nil {
						return "", err
					}
					fromClauseInfo = navigateFromClauseInfo
				} else {

					navigateFromClauseInfo.next, err = f.buildFromClauseInfo(expr, injectInfo)
					if err != nil {
						return "", err
					}
					navigateFromClauseInfo = navigateFromClauseInfo.next

				}
			case *sqlparser.ComparisonExpr:
				if navigateFromClauseInfo == nil {
					navigateFromClauseInfo, err = f.buildFromClauseInfoByComparisonExpr(expr, injectInfo)
					if err != nil {
						return "", err
					}
					fromClauseInfo = navigateFromClauseInfo
				} else {

					navigateFromClauseInfo.next, err = f.buildFromClauseInfoByComparisonExpr(expr, injectInfo)
					if err != nil {
						return "", err
					}
					navigateFromClauseInfo = navigateFromClauseInfo.next

				}
				// TODO: resolve colname
			default:
				panic(fmt.Sprintf("unimplemented: %T, see fromClauseType.Resolve", expr))
			}
		case *helper.InspectInfo:
			// suquery builder

			sql, err := f.ResolveQuery(node, injectInfo)
			if err != nil {
				return "", err
			}

			sourceTables = append(sourceTables, sql.Sql)

		default:
			panic(fmt.Sprintf("unimplemented: %T, see fromClauseType.Resolve", node))
		}
	}
	if fromClauseInfo == nil {
		ret := strings.Join(sourceTables, ", ")

		return ret, nil
	}

	return fromClauseInfo.String(), nil
}

//fromClauseType.buildFromClauseInfoByComparisonExpr.go

func (f *fromClauseType) buildFromClauseInfo(expr *sqlparser.FuncExpr, injectInfo *common.InjectInfo) (*fromClauseInfo, error) {
	fnName := strings.ToLower(expr.Name.String())
	switch fnName {
	case "left", "right":
		return f.buildFromClauseInfoByExpr(expr.Exprs[0], injectInfo, strings.ToUpper(fnName))
	case "leftouter":
		return f.buildFromClauseInfoByExpr(expr.Exprs[0], injectInfo, "LEFT OUTER")
	case "rightouter":
		return f.buildFromClauseInfoByExpr(expr.Exprs[0], injectInfo, "RIGHT OUTER")
	case "fullouter":
		return f.buildFromClauseInfoByExpr(expr.Exprs[0], injectInfo, "FULL OUTER")
	default:
		return nil, errors.NewParseError("unrecognized join type: %s", fnName)
	}
}

func (f *fromClauseType) buildFromClauseInfoByExpr(node sqlparser.SQLNode, injectInfo *common.InjectInfo, fnName string) (*fromClauseInfo, error) {
	switch node := node.(type) {
	case *sqlparser.AliasedExpr:
		return f.buildFromClauseInfoByExpr(node.Expr, injectInfo, fnName)
	case *sqlparser.ComparisonExpr:
		ret, err := f.buildFromClauseInfoByComparisonExpr(node, injectInfo)
		if err != nil {
			return nil, err
		}
		ret.joinType = fnName
		return ret, nil

	}
	panic(fmt.Sprintf("unimplemented: %T, see fromClauseType.buildFromClauseInfoByExpr", node))
}
