package from

import (
	"fmt"
	"strings"

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
func (f *fromClauseType) ResolveInSelect(nodes []any, injectInfo *common.InjectInfo) (string, error) {
	sourceTables := []string{}
	var navigateFromClauseInfo *fromClauseInfo
	var fromClauseInfo *fromClauseInfo
	var err error
	// injectInfo := &InjectInfo{
	// 	Args:        []common.ArgScaner{},
	// 	DyanmicArgs: args,
	// 	TextArgs:    texts,
	// 	Dialect:     dialect,
	// 	Scope:       scope,
	// 	Dict:        dict,
	// }
	for _, x := range nodes {
		switch x := x.(type) {
		case *helper.InspectInfo:
			sql, err := f.ResolveQuery(x, injectInfo)
			if err != nil {
				return "", err
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
							return "", err
						}
						fromClauseInfo = navigateFromClauseInfo
					} else {

						navigateFromClauseInfo.next, err = f.buildFromClauseInfo(n, injectInfo)
						if err != nil {
							return "", err
						}
						navigateFromClauseInfo = navigateFromClauseInfo.next

					}
				} else {
					tableName := n.Name.String()
					injectInfo.Dict.BuildByAliasTableName(injectInfo.Dialect, alias, tableName)
					if entityName, ok := injectInfo.Dict.Entities[strings.ToLower(tableName)]; ok {
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
				}
			case *sqlparser.ComparisonExpr:
				if navigateFromClauseInfo == nil {
					navigateFromClauseInfo, err = f.buildFromClauseInfoByComparisonExpr(n, injectInfo)
					if err != nil {
						return "", err
					}
					fromClauseInfo = navigateFromClauseInfo
				} else {

					navigateFromClauseInfo.next, err = f.buildFromClauseInfoByComparisonExpr(n, injectInfo)
					if err != nil {
						return "", err
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

		return ret, nil
	}

	return fromClauseInfo.String(), nil
}
