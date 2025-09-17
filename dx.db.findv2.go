package dx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (w *whereTypes) Find(item any) error {
	err := internal.Helper.AddrssertSinglePointerToSlice(item)
	if err != nil {
		return err
	}
	if w.err != nil {
		return w.err
	}
	whereStr, ars := w.getFilter()
	orderStr := ""
	if len(w.orders) > 0 {
		orderStr = strings.Join(w.orders, ",")
	}

	return w.db.findtWithFilterV2(item, w.ctx, w.sqlTx, whereStr, orderStr, w.limit, w.offset, true, ars...)
}

type findtWithFilterV2Key struct {
	DriverName string
	DbName     string
	eleType    reflect.Type
	filter     string
	order      string
	limit      uint64
	offset     uint64
}

func (db *DB) findtWithFilterV2(
	entity interface{},
	ctx context.Context,
	sqtTx *sql.Tx,
	filter string,
	orderStr string,
	limit,
	offset *uint64,
	resetLen bool,
	args ...interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Slice {
		return fmt.Errorf("%s is not slice", reflect.TypeOf(entity).String())
	}
	eleType := typ.Elem()
	key := findtWithFilterV2Key{
		eleType:    eleType,
		filter:     filter,
		order:      orderStr,
		DriverName: db.DriverName,
		DbName:     db.DbName,
	}
	//sql, err := db.buildBasicSqlFindItem(eleType, filter, order) //OnBuildSQLFirstItem(typ, db, filter)

	//key := db.DriverName + "://" + db.DbName + "/" + eleType.String() + "/buildBasicSqlFindItem/" + filter + "/" + orderStr
	if limit != nil {
		key.limit = *limit
		// key += fmt.Sprintf("/%d", *limit)
	}
	if offset != nil {
		key.offset = *offset
		//key += fmt.Sprintf("/%d", *offset)
	}
	sql, err := internal.OnceCall(key, func() (*types.SqlParse, error) {

		repoType, err := model.ModelRegister.GetModelByType(eleType)
		if err != nil {
			return nil, err
		}
		fieldsSelect := make([]string, len(repoType.Entity.Cols))
		for i, col := range repoType.Entity.Cols {
			fieldsSelect[i] = repoType.Entity.TableName + "." + col.Name + " AS " + col.Field.Name
		}
		sqlInfo := &types.SqlInfo{
			StrSelect: strings.Join(fieldsSelect, ","),
			From:      repoType.Entity.TableName,
			Limit:     limit,
			Offset:    offset,
		}

		if filter != "" {
			if err != nil {
				return nil, err
			}
			sqlInfo.StrWhere = filter

		}
		if orderStr != "" {
			sqlInfo.StrOrder = orderStr
			//sql += " ORDER BY " + compiler.Content
		}

		ret, err := compiler.GetSql(sqlInfo, db.DriverName)
		return ret, err
	})
	if err != nil {
		return err
	}
	if Options.ShowSql {
		fmt.Println(sql.Sql)
	}
	return db.fecthItems(entity, sql.Sql, ctx, sqtTx, resetLen, args...)

}
