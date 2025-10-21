package from

import (
	"strings"

	"github.com/vn-go/dx/dialect/types"

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

func (f *fromClauseType) ResolveQuery(scope common.AccessScope, dialect types.Dialect, info *helper.InspectInfo, dict *Dictionary) (*types.SqlParse, error) {
	var err error
	sqlInfo := &types.SqlInfo{
		ArgumentData: types.ArgsData{
			ArgWhere:   []any{},
			ArgsSelect: []any{},
			ArgJoin:    []any{},
			ArgGroup:   []any{},
			ArgHaving:  []any{},
			ArgOrder:   []any{},
			ArgSetter:  []any{},
		},
	}
	for k, v := range info.InspectData {
		if k == "from" {
			sqlInfo.From, sqlInfo.ArgumentData.ArgJoin, dict, err = f.Resolve(v.Nodes, scope, dialect, info.Texts, info.Args)
			if err != nil {
				return nil, err
			}
			return dialect.BuildSql(sqlInfo)
		} else if k == "select" {
			sqlInfo.From, sqlInfo.ArgumentData.ArgJoin, dict, err = f.ResolveInSelect(v.Nodes, scope, dialect, info.Texts, info.Args, dict)
			if err != nil {
				return nil, err
			}
			return dialect.BuildSql(sqlInfo)
		} else if f.KeyIsSubquery(strings.ToLower(k)) {
			k = strings.ToLower(strings.TrimSpace(k))
			info := v.Nodes[0].(*helper.InspectInfo)
			sql := &types.SqlParse{}
			strSql := []string{}
			//argsSql := internal.CompilerArgs{}
			for info != nil {
				sqlNext, err := f.ResolveQuery(scope, dialect, info, dict)

				if err != nil {
					return nil, err
				}
				strSql = append(strSql, sqlNext.Sql+" \n "+info.NextType)

				info = info.Next
			}
			sql.Sql = strings.Join(strSql, " \n ")

			sql.Sql = "(" + sql.Sql + ") " + dialect.Quote(k)
			return sql, nil
		}
	}
	return nil, nil

}


