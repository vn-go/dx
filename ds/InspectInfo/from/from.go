package from

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/common"
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
}

func (f *fromClauseType) Resolve(nodes []any, scope common.AccessScope, dialect types.Dialect, texts []string, args []any) (
	string, []any, *Dictionary, error) {

	dict := NewDictionary()
	var err error
	var navigateFromClauseInfo *fromClauseInfo
	var fromClauseInfo *fromClauseInfo
	injectInfo := &InjectInfo{
		Args:        []common.ArgScaner{},
		DyanmicArgs: args,
		TextArgs:    texts,
		Dialect:     dialect,
		Scope:       scope,
	}
	sourceTables := []string{}
	for key, node := range nodes {
		switch node := node.(type) {
		case *sqlparser.AliasedExpr:
			switch expr := node.Expr.(type) {
			case *sqlparser.ColName:
				dict.BuildByAliasTableName(dialect, node.As.String(), expr.Name.String())
				if entityName, ok := dict.Entities[strings.ToLower(expr.Name.String())]; ok {
					tableName := entityName.TableName
					if alias, ok := dict.TableAlias[strings.ToLower(tableName)]; ok {
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
			fmt.Println(key)
			sql, err := f.ResolveQuery(scope, dialect, node, dict)
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

		return ret, nil, dict, nil
	}

	return fromClauseInfo.String(), nil, dict, nil
}
//fromClauseType.buildFromClauseInfoByComparisonExpr.go


func (f *fromClauseType) buildFromClauseInfo(expr *sqlparser.FuncExpr, injectInfo *InjectInfo) (*fromClauseInfo, error) {
	panic("unimplemented")
}
