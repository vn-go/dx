package from

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/ds/helper"
	"github.com/vn-go/dx/sqlparser"
)

type fromClauseTypefromClauseTypeTable struct {
	tableName string
	alias     string
}

/*
this function try inspect the from clause and resolve the table name to a table object
*/
func (f *fromClauseType) ResolveInSelect(nodes []any, scope common.AccessScope, dialect types.Dialect, texts []string, args []any, dict *Dictionary) (string, []any, *Dictionary, error) {
	sourceTables := []string{}
	var navigateFromClauseInfo *fromClauseInfo
	var fromClauseInfo *fromClauseInfo
	var err error
	injectInfo := &InjectInfo{
		Args:        []common.ArgScaner{},
		DyanmicArgs: args,
		TextArgs:    texts,
		Dialect:     dialect,
		Scope:       scope,
		Dict:        dict,
	}
	for _, x := range nodes {
		switch x := x.(type) {
		case *helper.InspectInfo:
			sql, err := f.ResolveQuery(scope, dialect, x, dict)
			if err != nil {
				return "", nil, nil, err
			}
			sourceTables = append(sourceTables, sql.Sql)
		case *sqlparser.AliasedExpr:
			alias := x.As.String()
			switch n := x.Expr.(type) {
			case *sqlparser.FuncExpr:
				fnName := strings.ToLower(n.Name.String())
				if fnName == "left" || fnName == "right" || fnName == "leftouter" || fnName == "rightouter" {
					if navigateFromClauseInfo == nil {
						navigateFromClauseInfo, err = f.buildFromClauseInfo(n, injectInfo)
						if err != nil {
							return "", nil, nil, err
						}
						fromClauseInfo = navigateFromClauseInfo
					} else {

						navigateFromClauseInfo.next, err = f.buildFromClauseInfo(n, injectInfo)
						if err != nil {
							return "", nil, nil, err
						}
						navigateFromClauseInfo = navigateFromClauseInfo.next

					}
				} else {
					tableName := n.Name.String()
					dict.BuildByAliasTableName(dialect, alias, tableName)
					if entityName, ok := injectInfo.Dict.Entities[strings.ToLower(tableName)]; ok {
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
				}
			case *sqlparser.ComparisonExpr:
				if navigateFromClauseInfo == nil {
					navigateFromClauseInfo, err = f.buildFromClauseInfoByComparisonExpr(n, injectInfo)
					if err != nil {
						return "", nil, nil, err
					}
					fromClauseInfo = navigateFromClauseInfo
				} else {

					navigateFromClauseInfo.next, err = f.buildFromClauseInfoByComparisonExpr(n, injectInfo)
					if err != nil {
						return "", nil, nil, err
					}
					navigateFromClauseInfo = navigateFromClauseInfo.next

				}
			}
		default:
			panic(fmt.Sprintf("unsupport type %T, see fromClauseType.ResolveInSelect", x))
		}
	}
	if fromClauseInfo == nil {
		ret := strings.Join(sourceTables, ", ")

		return ret, nil, injectInfo.Dict, nil
	}

	return fromClauseInfo.String(), nil, injectInfo.Dict, nil
}
