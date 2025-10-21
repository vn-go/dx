package from

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
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
	Dict       Dictionary
}
type fromClauseInfo struct {
	left     string
	right    string
	on       string
	joinType string
	next     *fromClauseInfo
	top      *fromClauseInfo
}

type InjectInfo struct {
	Args        []common.ArgScaner
	DyanmicArgs []any
	TextArgs    []string
	Dialect     types.Dialect
	Scope       common.AccessScope
	Dict        *Dictionary
}

func (f *fromClauseType) Resolve(nodes []any, scope common.AccessScope, dialect types.Dialect, texts []string, args []any) (
	string, []any, *Dictionary, error) {

	var err error
	var navigateFromClauseInfo *fromClauseInfo
	var fromClauseInfo *fromClauseInfo
	injectInfo := &InjectInfo{
		Args:        []common.ArgScaner{},
		DyanmicArgs: args,
		TextArgs:    texts,
		Dialect:     dialect,
		Scope:       scope,
		Dict:        NewDictionary(),
	}
	sourceTables := []string{}
	for _, node := range nodes {
		switch node := node.(type) {
		case *sqlparser.AliasedExpr:
			switch expr := node.Expr.(type) {
			case *sqlparser.ColName:
				injectInfo.Dict.BuildByAliasTableName(dialect, node.As.String(), expr.Name.String())
				if entityName, ok := injectInfo.Dict.Entities[strings.ToLower(expr.Name.String())]; ok {
					tableName := entityName.TableName
					if alias, ok := injectInfo.Dict.TableAlias[strings.ToLower(tableName)]; ok {
						if alias != "" {
							sourceTables = append(sourceTables, dialect.Quote(tableName)+" "+dialect.Quote(alias))
						} else {
							sourceTables = append(sourceTables, dialect.Quote(tableName))
						}
					} else {
						sourceTables = append(sourceTables, dialect.Quote(tableName))
					}
				}

				// TODO: resolve colname
			case *sqlparser.FuncExpr:
				if navigateFromClauseInfo == nil {
					navigateFromClauseInfo, err = f.buildFromClauseInfo(expr, injectInfo)
					if err != nil {
						return "", nil, nil, err
					}
					fromClauseInfo = navigateFromClauseInfo
				} else {

					navigateFromClauseInfo.next, err = f.buildFromClauseInfo(expr, injectInfo)
					if err != nil {
						return "", nil, nil, err
					}
					navigateFromClauseInfo = navigateFromClauseInfo.next

				}
			case *sqlparser.ComparisonExpr:
				if navigateFromClauseInfo == nil {
					navigateFromClauseInfo, err = f.buildFromClauseInfoByComparisonExpr(expr, injectInfo)
					if err != nil {
						return "", nil, nil, err
					}
					fromClauseInfo = navigateFromClauseInfo
				} else {

					navigateFromClauseInfo.next, err = f.buildFromClauseInfoByComparisonExpr(expr, injectInfo)
					if err != nil {
						return "", nil, nil, err
					}
					navigateFromClauseInfo = navigateFromClauseInfo.next

				}
				// TODO: resolve colname
			default:
				panic(fmt.Sprintf("unimplemented: %T, see fromClauseType.Resolve", expr))
			}
		case *helper.InspectInfo:
			// suquery builder

			sql, err := f.ResolveQuery(scope, dialect, node, injectInfo.Dict)
			if err != nil {
				return "", nil, nil, err
			}

			sourceTables = append(sourceTables, sql.Sql)

		default:
			panic(fmt.Sprintf("unimplemented: %T, see fromClauseType.Resolve", node))
		}
	}
	if fromClauseInfo == nil {
		ret := strings.Join(sourceTables, ", ")

		return ret, nil, injectInfo.Dict, nil
	}

	return fromClauseInfo.String(), nil, injectInfo.Dict, nil
}

//fromClauseType.buildFromClauseInfoByComparisonExpr.go

func (f *fromClauseType) buildFromClauseInfo(expr *sqlparser.FuncExpr, injectInfo *InjectInfo) (*fromClauseInfo, error) {
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

func (f *fromClauseType) buildFromClauseInfoByExpr(node sqlparser.SQLNode, injectInfo *InjectInfo, fnName string) (*fromClauseInfo, error) {
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
