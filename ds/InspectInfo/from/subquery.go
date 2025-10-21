package from

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"

	"github.com/vn-go/dx/ds/InspectInfo/selectors"
	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/ds/helper"
)

var mapSqlKeyword = map[string]bool{
	"select":  true,
	"from":    true,
	"where":   true,
	"union":   true,
	"orderby": true,
}

func (f *fromClauseType) KeyIsSubquery(key string) bool {
	if _, ok := mapSqlKeyword[key]; !ok {
		return true
	}
	return false
}

func (f *fromClauseType) ResolveQuery(info *helper.InspectInfo, injectInfo *common.InjectInfo) (*types.SqlParse, error) {
	var err error
	sqlInfo := &types.SqlInfo{}

	dialect := injectInfo.Dialect
	//var selectorNode *helper.SqlNode = nil
	for k, v := range info.InspectData {
		if k == "from" {
			sqlInfo.From, err = f.Resolve(v.Nodes, injectInfo)
			if err != nil {
				return nil, err
			}
			break

			//return dialect.BuildSql(sqlInfo)
		} else if k == "select" {
			//selectorNode = &v
			sqlInfo.From, err = f.ResolveInSelect(v.Nodes, injectInfo)
			if err != nil {
				return nil, err
			}
			break

		} else if f.KeyIsSubquery(strings.ToLower(k)) {
			/*
			  begin build subquery
			*/
			k = strings.ToLower(strings.TrimSpace(k))
			info := v.Nodes[0].(*helper.InspectInfo)
			sql := &types.SqlParse{}
			strSql := []string{}
			injectInfo.OuputFieldsInSelectorStack.Push(injectInfo.OuputFieldsInSelector)
			injectInfo.OuputFieldsInSelector = []string{}
			var OuputFieldsInSelector []string
			for info != nil {
				sqlNext, err := f.ResolveQuery(info, injectInfo)
				if err != nil {
					return nil, err
				}
				if OuputFieldsInSelector == nil {
					// for union query jut get first select fields
					OuputFieldsInSelector = injectInfo.OuputFieldsInSelector
				}
				strSql = append(strSql, sqlNext.Sql+"\n "+info.NextType)

				info = info.Next // move to next node, the next node is the next union query
				// union query was complied with query->next(query->next(query->next(nil)))
			}
			injectInfo.Dict.FieldMap = map[string]common.DictionaryItem{} // reset field map for subquery

			for _, col := range OuputFieldsInSelector {
				key := fmt.Sprintf("%s.%s", k, strings.ToLower(col))
				injectInfo.Dict.FieldMap[key] = common.DictionaryItem{
					Content: injectInfo.Dialect.Quote(k, col),
					Alias:   col,
				}
			}
			sql.Sql = strings.Join(strSql, " \n ")

			sqlInfo.From = "(" + sql.Sql + ") " + dialect.Quote(k)
			if OuputFieldsInSelector, ok := injectInfo.OuputFieldsInSelectorStack.Pop(); ok {
				injectInfo.OuputFieldsInSelector = OuputFieldsInSelector
			}
			return sql, nil
			/*
				end build subquery
				Subquery was neasted query, so we need to return to up level of tree expression for next compiling
			*/
		}
	}
	if selectData, ok := info.InspectData["select"]; ok {
		sqlInfo.StrSelect, err = selectors.Selector.Resolve(selectData.Nodes, injectInfo, f.ResolveQuery)

		if err != nil {
			return nil, err
		}
	}
	return dialect.BuildSql(sqlInfo)

}
