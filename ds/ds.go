package ds

import (
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/InspectInfo/from"
	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/internal"

	"github.com/vn-go/dx/ds/helper"
)

func Compile(dialect types.Dialect, query string, args ...any) (*types.SqlParse, error) {
	var err error

	info, texts, err := helper.Inspect(query)
	if err != nil {
		return nil, err
	}
	info.Args = args
	info.Texts = texts
	sql := &types.SqlParse{}
	strSql := []string{}
	//argsSql := internal.CompilerArgs{}
	injectInfo := &common.InjectInfo{
		Args:        []common.ArgScaner{},
		DyanmicArgs: args,
		TextArgs:    texts,
		Dialect:     dialect,

		Dict:                       common.NewDictionary(),
		OuputFieldsInSelector:      []string{},
		OuputFieldsInSelectorStack: internal.Stack[[]string]{},
	}
	for info != nil {
		sqlNext, err := from.FromClause.ResolveQuery(info, injectInfo)
		if err != nil {
			return nil, err
		}
		strSql = append(strSql, sqlNext.Sql+" \n "+info.NextType)

		info = info.Next
	}
	sql.Sql = strings.Join(strSql, " \n ")
	sql.Sql = strings.Trim(sql.Sql, " \n")
	return sql, nil
}
