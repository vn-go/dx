package mssql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
)

func mssqlBuilSql(info types.SqlInfo) (*types.SqlParse, error) {
	if info.SqlType == types.SQL_SELECT {
		return mssqlBuilSqlSelect(info)
	}
	if info.SqlType == types.SQL_DELETE {
		return mssqlBuilSqlDelete(info)
	}
	if info.SqlType == types.SQL_UPDATE {
		return mssqlBuildSqlUpdate(info)
	}
	panic(fmt.Sprintf("not support %s, see 'mssqlBuilSql' file %s", info.SqlType, `dialect\mssql\BuildSql.go`))
}
func mssqlBuildSqlUpdate(info types.SqlInfo) (*types.SqlParse, error) {
	var sb strings.Builder
	ret := &types.SqlParse{
		ArgIndex: []reflect.StructField{},
	}
	if strFrom, ok := info.From.(string); ok {
		_, err := sb.WriteString("UPDATE " + strFrom)
		if err != nil {
			return nil, err
		}
	} else {
		panic(fmt.Sprintf("not support %s with from %T, see file %s", info.SqlType, info.From, `dialect\mysql\BuildSql.go`))
	}
	_, err := sb.WriteString(" SET  " + info.StrSetter)
	if err != nil {
		return nil, err
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgWhere)
	if info.StrWhere != "" {
		_, err := sb.WriteString(" WHERE " + info.StrWhere)
		if err != nil {
			return nil, err
		}
	}
	ret.Sql = sb.String()
	return ret, nil

}
func mssqlBuilSqlDelete(info types.SqlInfo) (*types.SqlParse, error) {
	var sb strings.Builder
	ret := &types.SqlParse{
		ArgIndex: []reflect.StructField{},
	}
	if strFrom, ok := info.From.(string); ok {
		_, err := sb.WriteString("DELETE FROM " + strFrom)
		if err != nil {
			return nil, err
		}
	} else {
		panic(fmt.Sprintf("not support %s with from %T, see file %s", info.SqlType, info.From, `dialect\mysql\BuildSql.go`))
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgWhere)
	if info.StrWhere != "" {
		_, err := sb.WriteString(" WHERE " + info.StrWhere)
		if err != nil {
			return nil, err
		}
	}
	ret.Sql = sb.String()
	return ret, nil

}
func mssqlBuilSqlSelect(info types.SqlInfo) (*types.SqlParse, error) {
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
		ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgJoin)
	case types.SqlInfo:
		inner, err := mssqlBuilSqlSelect(v)
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
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgOrder)
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
	if info.UnionNext != nil {
		sqlParse, err := mssqlBuilSqlSelect(*info.UnionNext)
		if err != nil {
			return nil, err
		}
		_, err = sb.WriteString(" " + info.UnionType + " ")
		if err != nil {
			return nil, err
		}
		_, err = sb.WriteString(sqlParse.Sql)
		if err != nil {
			return nil, err
		}
		ret.ArgIndex = append(ret.ArgIndex, sqlParse.ArgIndex...)
	}
	ret.Sql = sb.String()
	return ret, nil
}
func (mssql *mssqlDialect) BuildSql(info *types.SqlInfo) (*types.SqlParse, error) {

	return internal.OnceCall(info.GetKey(), func() (*types.SqlParse, error) {
		return mssqlBuilSql(*info)
	})
}
func (mssql *mssqlDialect) BuildSqlNoCache(info *types.SqlInfo) (*types.SqlParse, error) {

	return mssqlBuilSql(*info)
}
