package ds

import (
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/InspectInfo/from"
	"github.com/vn-go/dx/ds/common"

	"github.com/vn-go/dx/ds/helper"
)

func Compile(scope common.AccessScope, dialect types.Dialect, query string, args ...any) (*types.SqlParse, error) {
	var err error
	dict := from.NewDictionary()
	info, texts, err := helper.Inspect(query)
	if err != nil {
		return nil, err
	}
	info.Args = args
	info.Texts = texts
	sql := &types.SqlParse{}
	strSql := []string{}
	//argsSql := internal.CompilerArgs{}
	for info != nil {
		sqlNext, err := from.FromClause.ResolveQuery(scope, dialect, info, dict)
		if err != nil {
			return nil, err
		}
		strSql = append(strSql, sqlNext.Sql+" \n "+info.NextType)

		info = info.Next
	}
	sql.Sql = strings.Join(strSql, " \n ")
	return sql, nil
}
