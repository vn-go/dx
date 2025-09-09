package dx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

func (db *DB) firstWithFilter(entity interface{}, filter string, ctx context.Context, sqlTx *sql.Tx, args ...interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, err := db.buildBasicSqlFirstItem(typ, filter) //OnBuildSQLFirstItem(typ, db, filter)
	if err != nil {
		return err
	}
	return db.ExecToItem(entity, sql, ctx, sqlTx, args...)

}
func (db *DB) firstWithFilterV2(entity interface{}, filter string, ctx context.Context, sqlTx *sql.Tx, args ...interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, err := db.buildBasicSqlFirstItemV2(typ, filter) //OnBuildSQLFirstItem(typ, db, filter)
	if err != nil {
		return err
	}
	return db.ExecToItem(entity, sql, ctx, sqlTx, args...)

}
func (db *DB) findtWithFilter(
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

	//sql, err := db.buildBasicSqlFindItem(eleType, filter, order) //OnBuildSQLFirstItem(typ, db, filter)

	key := db.DriverName + "://" + db.DbName + "/" + eleType.String() + "/buildBasicSqlFindItem/" + filter + "/" + orderStr
	if limit != nil {
		key += fmt.Sprintf("/%d", *limit)
	}
	if offset != nil {
		key += fmt.Sprintf("/%d", *offset)
	}
	sql, err := internal.OnceCall(key, func() (string, error) {

		repoType, err := model.ModelRegister.GetModelByType(eleType)
		if err != nil {
			return "", err
		}
		tableName := repoType.Entity.TableName
		compiler, err := expr.CompileJoin(tableName, db.DB)
		if err != nil {
			return "", err
		}
		tableName = compiler.Content
		columns := repoType.Entity.Cols

		fieldsSelect := make([]string, len(columns))
		for i, col := range columns {
			fieldsSelect[i] = repoType.Entity.TableName + "." + col.Field.Name + " AS " + col.Field.Name
		}
		compiler.Context.Purpose = expr.BUILD_SELECT
		err = compiler.BuildSelectField(strings.Join(fieldsSelect, ", "))
		if err != nil {
			return "", err
		}
		strField := compiler.Content

		sql := fmt.Sprintf("SELECT %s FROM %s", strField, tableName)
		if filter != "" {
			compiler.Context.Purpose = expr.BUILD_WHERE
			err = compiler.BuildWhere(filter)
			if err != nil {
				return "", err
			}
			sql += " WHERE " + compiler.Content
		}
		if orderStr != "" {
			compiler.Context.Purpose = expr.BUILD_ORDER
			err = compiler.BuildSortField(orderStr)
			if err != nil {
				return "", err
			}
			sql = compiler.Context.Dialect.LimitAndOffset(sql, limit, offset, compiler.Content)
			//sql += " ORDER BY " + compiler.Content
		} else {
			sql = compiler.Context.Dialect.LimitAndOffset(sql, limit, offset, "")
		}

		return sql, nil
	})
	if err != nil {
		return err
	}

	return db.fecthItems(entity, sql, ctx, sqtTx, resetLen, args...)

}
