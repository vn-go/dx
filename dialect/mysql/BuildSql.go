package mysql

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
)

func mySqlbuildSQL(info types.SqlInfo) (*types.SqlParse, error) {
	if info.SqlType == types.SQL_SELECT {
		return mySqlbuildSqlSelect(info)
	}
	if info.SqlType == types.SQL_DELETE {
		return myqlbuildSqlDelete(info)
	}
	if info.SqlType == types.SQL_UPDATE {
		return myqlbuildSqlUpdate(info)
	}
	panic(fmt.Sprintf("not support %s, see file %s", info.SqlType, `dialect\mysql\BuildSql.go`))
}
func myqlbuildSqlUpdate(info types.SqlInfo) (*types.SqlParse, error) {
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
func myqlbuildSqlDelete(info types.SqlInfo) (*types.SqlParse, error) {
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
func mySqlbuildSqlSelect(info types.SqlInfo) (*types.SqlParse, error) {
	var sb strings.Builder
	ret := &types.SqlParse{
		ArgIndex: []reflect.StructField{},
	}

	// SELECT
	if info.StrSelect == "" {
		_, err := sb.WriteString("SELECT *")
		if err != nil {
			return nil, err
		}
	} else {
		_, err := sb.WriteString("SELECT " + info.StrSelect)
		if err != nil {
			return nil, err
		}
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgsSelect)
	// FROM
	switch v := info.From.(type) {
	case string:
		if v != "" {
			_, err := sb.WriteString(" FROM " + v)
			if err != nil {
				return nil, err
			}
		}
	case types.SqlInfo:
		inner, err := mySqlbuildSqlSelect(v)
		if err != nil {
			return nil, err
		}
		ret.ArgIndex = append(ret.ArgIndex, inner.ArgIndex...)
		_, err = sb.WriteString(" FROM (" + inner.Sql + ") AS T")
		if err != nil {
			return nil, err
		}
	default:
		// nothing
	}

	// WHERE
	if info.StrWhere != "" {
		_, err := sb.WriteString(" WHERE " + info.StrWhere)
		if err != nil {
			return nil, err
		}
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgWhere)
	// GROUP BY
	if info.StrGroupBy != "" {
		_, err := sb.WriteString(" GROUP BY " + info.StrGroupBy)
		if err != nil {
			return nil, err
		}
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgGroup)
	// HAVING
	if info.StrHaving != "" {
		_, err := sb.WriteString(" HAVING " + info.StrHaving)
		if err != nil {
			return nil, err
		}
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgHaving)
	// ORDER BY
	if info.StrOrder != "" {
		_, err := sb.WriteString(" ORDER BY " + info.StrOrder)
		if err != nil {
			return nil, err
		}
	}
	ret.ArgIndex = append(ret.ArgIndex, info.FieldArs.ArgGroup)
	// LIMIT + OFFSET (chuẩn MySQL)
	if info.Limit != nil {
		_, err := sb.WriteString(fmt.Sprintf(" LIMIT %d", *info.Limit))
		if err != nil {
			return nil, err
		}
		if info.Offset != nil {
			_, err := sb.WriteString(fmt.Sprintf(" OFFSET %d", *info.Offset))
			if err != nil {
				return nil, err
			}
		}
	}
	if info.UnionNext != nil {
		sqlParse, err := mySqlbuildSqlSelect(*info.UnionNext)
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
func (mssql *mySqlDialect) BuildSql(info *types.SqlInfo) (*types.SqlParse, error) {
	return internal.OnceCall(info.GetKey(), func() (*types.SqlParse, error) {
		return mySqlbuildSQL(*info)
	})
}
