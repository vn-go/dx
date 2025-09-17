package postgres

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
)

func postgresBuilSql(info types.SqlInfo) (*types.SqlParse, error) {
	if info.SqlType == types.SQL_SELECT {
		return postgresBuilSqlSelect(info)
	}
	if info.SqlType == types.SQL_DELETE {
		return postgresBuilSqlDelete(info)
	}
	if info.SqlType == types.SQL_UPDATE {
		return postresBuildSqlUpdate(info)
	}
	panic(fmt.Sprintf("not support %s, see 'postgresBuilSql' file %s", info.SqlType, `dialect\postgres\BuildSql.go`))
}
func postresBuildSqlUpdate(info types.SqlInfo) (*types.SqlParse, error) {
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
func postgresBuilSqlDelete(info types.SqlInfo) (*types.SqlParse, error) {
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
		panic(fmt.Sprintf("not support %s with from %T, see file %s", info.SqlType, info.From, `dialect\postgres\BuildSql.go`))
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
func postgresBuilSqlSelect(info types.SqlInfo) (*types.SqlParse, error) {
	var sb strings.Builder
	ret := &types.SqlParse{
		ArgIndex: []reflect.StructField{},
	}

	// SELECT
	if info.StrSelect == "" {
		sb.WriteString("SELECT *")
	} else {
		sb.WriteString("SELECT " + info.StrSelect)
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgsSelect)
	// FROM
	switch v := info.From.(type) {
	case string:
		if v != "" {
			sb.WriteString(" FROM " + v)
			ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgJoin)
		}
	case types.SqlInfo:
		inner, err := postgresBuilSqlSelect(v)
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
	// LIMIT + OFFSET (chuẩn MySQL)
	if info.Limit != nil {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", *info.Limit))
		if info.Offset != nil {
			sb.WriteString(fmt.Sprintf(" OFFSET %d", *info.Offset))
		}
	}
	if info.UnionNext != nil {
		sqlParse, err := postgresBuilSqlSelect(*info.UnionNext)
		if err != nil {
			return nil, err
		}
		sb.WriteString(" " + info.UnionType + " ")
		sb.WriteString(sqlParse.Sql)
		ret.ArgIndex = append(ret.ArgIndex, sqlParse.ArgIndex...)
	}
	ret.Sql = sb.String()
	return ret, nil
}
func (mssql *postgresDialect) BuildSql(info *types.SqlInfo) (*types.SqlParse, error) {
	return internal.OnceCall(info.GetKey(), func() (*types.SqlParse, error) {
		return postgresBuilSql(*info)
	})
}
