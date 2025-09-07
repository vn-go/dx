package mssql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
)

func buildSQLmssql(info types.SqlInfo) (string, error) {
	var sb strings.Builder

	// SELECT
	if info.StrSelect == "" {
		sb.WriteString("SELECT *")
	} else {
		sb.WriteString("SELECT " + info.StrSelect)
	}

	// FROM
	switch v := info.From.(type) {
	case string:
		if v != "" {
			sb.WriteString(" FROM " + v)
		}
	case types.SqlInfo:
		inner, err := buildSQLmssql(v)
		if err != nil {
			return "", err
		}
		sb.WriteString(" FROM (" + inner + ") AS T")
	default:
		// nothing
	}

	// WHERE
	if info.StrWhere != "" {
		sb.WriteString(" WHERE " + info.StrWhere)
	}

	// GROUP BY
	if info.StrGroupBy != "" {
		sb.WriteString(" GROUP BY " + info.StrGroupBy)
	}

	// HAVING
	if info.StrHaving != "" {
		sb.WriteString(" HAVING " + info.StrHaving)
	}

	// ORDER BY
	if info.StrOrder != "" {
		sb.WriteString(" ORDER BY " + info.StrOrder)
	}

	// LIMIT + OFFSET (chuẩn MySQL)
	if info.Limit != nil {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", *info.Limit))
		if info.Offset != nil {
			sb.WriteString(fmt.Sprintf(" OFFSET %d", *info.Offset))
		}
	}

	return sb.String(), nil
}
func (mssql *mssqlDialect) BuildSql(info *types.SqlInfo) (string, error) {

	return internal.OnceCall(info.GetKey(), func() (string, error) {
		return buildSQLmssql(*info)
	})
}
