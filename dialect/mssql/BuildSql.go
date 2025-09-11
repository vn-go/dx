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
	/*
				SELECT [T1].[id] [ID],
		       [T1].[user_id] [UserId],
		       [T1].[email] [Email],
		       [T1].[phone] [Phone],
		       [T1].[username] [Username],
		       [T1].[hash_password] [HashPassword],
		       [T1].[record_id] [RecordID],
		       [T1].[created_at] [CreatedAt],
		       [T1].[updated_at] [UpdatedAt],
		       [T1].[description] [Description]
		FROM [users] [T1]
		ORDER BY (SELECT NULL)
		OFFSET 1000 ROWS FETCH NEXT 100 ROWS ONLY;
	*/
	if info.Offset != nil {
		limit := uint64(0)
		if info.Limit != nil {
			limit = *info.Limit
		}
		if info.StrOrder == "" {
			sb.WriteString(" ORDER BY (SELECT NULL)")
		}
		sb.WriteString(fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", *info.Offset, limit))
	}

	ret.Sql = sb.String()
	return ret, nil
}
func (mssql *mssqlDialect) BuildSql(info *types.SqlInfo) (*types.SqlParse, error) {

	return internal.OnceCall(info.GetKey(), func() (*types.SqlParse, error) {
		return buildSQLmssql(*info)
	})
}
