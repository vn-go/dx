package sql

import (
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type simpleSql struct {
	from    string
	where   string
	selects string
	sort    string
}

func (s *smarty) simple(simpleQuery string) (string, error) {
	ret := &simpleSql{}
	str, err := internal.Helper.QuoteExpression2(simpleQuery)
	if err != nil {
		return "", err
	}

	tk := sqlparser.NewStringTokenizer("select " + str)
	stm, err := sqlparser.ParseNext(tk)

	if err != nil {
		panic(err)
	}
	selectStm := stm.(*sqlparser.Select)
	ret.from = smartier.from(selectStm)

	ret.selects = smartier.selectors(selectStm)

	ret.where = smartier.where(selectStm)
	return ret.String(), nil
}

func (sql *simpleSql) String() string {
	query := "SELECT " + sql.selects

	if sql.from != "" {
		query += " FROM " + sql.from
	}
	if sql.where != "" {
		query += " WHERE " + sql.where
	}
	if sql.sort != "" {
		query += " ORDER BY " + sql.sort
	}

	return query
}
