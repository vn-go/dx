package mssql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
)

func buildSQLmssql(info types.SqlInfo) (*types.SqlParse, error) {
	var sb strings.Builder
	ret := &types.SqlParse{
		ArgIndex: []reflect.StructField{},
	}
	// SELECT
	if info.StrSelect == "" {
		if info.Offset == nil && info.Limit != nil {
			sb.WriteString(fmt.Sprintf("SELECT TOP %d *", *info.Limit))
		} else {
			sb.WriteString("SELECT *")
		}

	} else {
		if info.Offset == nil && info.Limit != nil {
			sb.WriteString(fmt.Sprintf("SELECT TOP %d %s", *info.Limit, info.StrSelect))
		} else {
			sb.WriteString("SELECT " + info.StrSelect)
		}

	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgsSelect)
	// FROM
	switch v := info.From.(type) {
	case string:
		if v != "" {
			sb.WriteString(" FROM " + v)
		}

	case types.SqlInfo:
		inner, err := buildSQLmssql(v)
		if err != nil {
			return nil, err

		}
		ret.ArgIndex = append(ret.ArgIndex, inner.ArgIndex...)
		sb.WriteString(" FROM (" + inner.Sql + ") AS T")
	default:
		// nothing
	}

	// WHERE
	if info.StrWhere != "" {
		sb.WriteString(" WHERE " + info.StrWhere)
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgWhere)

	// GROUP BY
	if info.StrGroupBy != "" {
		sb.WriteString(" GROUP BY " + info.StrGroupBy)
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgGroup)

	// HAVING
	if info.StrHaving != "" {
		sb.WriteString(" HAVING " + info.StrHaving)
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgHaving)

	// ORDER BY
	if info.StrOrder != "" {
		sb.WriteString(" ORDER BY " + info.StrOrder)
	}

	// LIMIT + OFFSET (chuẩn MySQL)
	if info.Offset != nil {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", *info.Limit))
		if info.Offset != nil {
			sb.WriteString(fmt.Sprintf(" OFFSET %d", *info.Offset))
		}
	}
	ret.Sql = sb.String()
	return ret, nil
}
func (mssql *mssqlDialect) BuildSql(info *types.SqlInfo) (*types.SqlParse, error) {

	return internal.OnceCall(info.GetKey(), func() (*types.SqlParse, error) {
		return buildSQLmssql(*info)
	})
}
