package tenantDB

import (
	"context"
	"fmt"
	"reflect"
)

type query struct {
	db         *TenantDB
	qrInstance interface{}
	sql        string
	args       []interface{}
	Err        error
}
type OnNewQr func() interface{}

var OnNewQrFn OnNewQr
var OnFrom OnExpr

type onLiterals = func(string) interface{}

var OnLiterals onLiterals

func (db *TenantDB) Lit(str string) interface{} {
	return OnLiterals(str)
}

type AliasResult struct {
	Alias string
	Err   error
}

func (db *TenantDB) From(table interface{}) *query {
	if table == nil {
		return &query{
			Err: fmt.Errorf("table is nil"),
		}
	}
	var err error
	tableName := ""
	if alais, ok := table.(AliasResult); ok {
		if alais.Err != nil {
			return &query{
				Err: alais.Err,
			}
		}
		tableName = alais.Alias
	} else {
		typ := reflect.TypeOf(table)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		tableName, err = ModelRegistry.getTableName(typ)
		if err != nil {
			return &query{
				Err: err,
			}
		} else if tableName == "" {
			return &query{
				Err: fmt.Errorf("table name was not fonud for type %s", typ.String()),
			}
		}

	}

	ret := createQr(db)
	ret.qrInstance = OnFrom(ret.qrInstance, tableName)

	return ret
}
func createQr(db *TenantDB) *query {
	ret := &query{
		db: db,
	}
	ret.qrInstance = OnNewQrFn()
	return ret
}

type selectExpr struct {
	Expr string
	Args []interface{}
}
type litOfString struct {
	val string
}

/*
Example:

	Select("id", "name", "age") <-- no args

	Select("id", "name", "age", 123) <-- with args

	In case arg is string select can not recognize it as parameter and will be treated as string literal
	Select("concat(firstName, ?,lastName)",Lit(" ")) <-- with args
*/
type onQrSelect func(qrInstance interface{}, exprsAndArgs ...interface{}) interface{}

var OnQrSelect onQrSelect

func (q *query) Select(exprsAndArgs ...interface{}) *query {

	q.qrInstance = OnQrSelect(q.qrInstance, exprsAndArgs...)
	return q
}

type onJoin = func(qrInstance interface{}, table string, onExpr string, args ...interface{}) interface{}

var OnInnerJoin onJoin
var OnLeftJoin onJoin
var OnRightJoin onJoin
var OnFullJoin onJoin

func (q *query) InnerJoin(table string, onExpr string, args ...interface{}) *query {
	q.qrInstance = OnInnerJoin(q.qrInstance, table, onExpr, args...)
	return q
}
func (q *query) LeftJoin(table interface{}, onExpr string, args ...interface{}) *query {
	tableName := ""
	var err error
	if strTableName, ok := table.(string); ok {
		tableName = strTableName
	} else {
		typ := reflect.TypeOf(table)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		tableName, err = ModelRegistry.getTableName(typ)
		if err != nil {
			return &query{
				Err: err,
			}
		}
		if tableName == "" {
			return &query{
				Err: fmt.Errorf("table name was not fonud for type %s", typ.String()),
			}
		}
	}

	q.qrInstance = OnLeftJoin(q.qrInstance, tableName, onExpr, args...)
	return q
}

type OnExpr = func(qrInstance interface{}, expr string, args ...interface{}) interface{}

var OnWhere OnExpr

func (q *query) Where(expr string, args ...interface{}) *query {
	q.qrInstance = OnWhere(q.qrInstance, expr, args...)
	return q
}

type onQueryArgs = func(qrInstance interface{}, args ...interface{}) interface{}

var OnOrderBy onQueryArgs

func (q *query) OrderBy(args ...interface{}) *query {
	q.qrInstance = OnOrderBy(q.qrInstance, args...)
	return q
}

var OnGroupBy onQueryArgs

func (q *query) GroupBy(args ...interface{}) *query {
	q.qrInstance = OnGroupBy(q.qrInstance, args...)
	return q
}

type onOffsetLimit = func(qrInstance interface{}, offset, limit int) interface{}

var OnOffsetLimit onOffsetLimit

func (q *query) OffsetLimit(offset, limit int) *query {
	q.qrInstance = OnOffsetLimit(q.qrInstance, offset, limit)
	return q
}

var OnHaving OnExpr

func (q *query) Having(expr string, args ...interface{}) *query {
	q.qrInstance = OnHaving(q.qrInstance, expr, args...)
	return q
}

type onBuildSql = func(qrInstance interface{}, db *TenantDB) (string, []interface{}, error)

var OnBuildSql onBuildSql

func (q *query) BuildSql() (string, []interface{}) {
	if q.sql != "" {
		return q.sql, q.args
	}
	sql, args, err := OnBuildSql(q.qrInstance, q.db)
	// fmt.Println("BuildSql", sql, args, err)
	if err != nil {
		q.Err = err
		return "", nil
	}

	q.sql = sql
	q.args = args
	return sql, args
}

type onBuildSQL func(typ reflect.Type, db *TenantDB, filter string) (string, error)

var OnBuildSQL onBuildSQL

/*
Example:

	r.db.Find(&users, "email = ? OR username = ?", identifier, identifier)
*/
func (db *TenantDB) Find(entity interface{}, filter string, args ...interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, err := OnBuildSQL(typ, db, filter)
	if err != nil {
		return err
	}
	return db.ExecToArray(entity, sql, args...)

}

//ype onBuildSQLFirstItem func(typ reflect.Type, db *TenantDB, filter string) (string, error)

/*
Get first item by filter
@entity
@fiter
@args

	Example:
			db.First(&model,"id={1}",1)
*/
func (db *TenantDB) First(entity interface{}, args ...interface{}) error {
	if len(args) == 0 {
		return db.firstWithNoFilter(entity)
	} else if len(args) >= 2 {
		if filter, ok := args[0].(string); ok {
			return db.firstWithFilter(entity, filter, args[1:]...)
		} else {
			return fmt.Errorf("first with filter: filter must be string")
		}

	} else {
		return fmt.Errorf("first with filter: filter must be string")
	}
}
func (db *TenantDB) firstWithNoFilter(entity interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, _, _, err := OnBuildSQLFirstItemNoFilter(typ, db)

	if err != nil {
		return err
	}
	return db.ExecToItem(entity, sql)

}
func (db *TenantDB) firstWithFilter(entity interface{}, filter string, args ...interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, err := OnBuildSQLFirstItem(typ, db, filter)
	if err != nil {
		return err
	}
	return db.ExecToItem(entity, sql, args...)

}

type onBuildSQLFirstItemNoFilter func(typ reflect.Type, db *TenantDB) (string, string, [][]int, error)

var OnBuildSQLFirstItemNoFilter onBuildSQLFirstItemNoFilter

func (db *TenantDB) FirstWithContext(context context.Context, entity interface{}, args ...interface{}) error {
	if len(args) == 0 {
		return db.FirstWithContextNoFilter(context, entity)
	} else if len(args) >= 2 {
		if filter, ok := args[0].(string); ok {
			return db.FirstWithContextAndFilter(context, entity, filter, args[1:]...)
		} else {
			return fmt.Errorf("first with context and filter: filter must be string")
		}

	} else {
		return fmt.Errorf("first with context and filter: filter must be string")
	}
}
func (db *TenantDB) FirstWithContextNoFilter(context context.Context, entity interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, _, _, err := OnBuildSQLFirstItemNoFilter(typ, db)

	val := reflect.ValueOf(entity)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if err != nil {
		return err
	}
	return db.ExecToItemWithContext(context, entity, sql)

}
func (db *TenantDB) FirstWithContextAndFilter(context context.Context, entity interface{}, filter string, args ...interface{}) error {
	typ := reflect.TypeOf(entity)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	sql, err := query.BuildBasicSqlFirstItem(typ, db, filter)
	if err != nil {
		return err
	}
	return db.ExecToItemWithContext(context, entity, sql, args...)

}

type whereExpr struct {
	expr string
	args []interface{}
	db   *TenantDB
}

func (db *TenantDB) Where(expr string, args ...interface{}) *whereExpr {
	return &whereExpr{
		db:   db,
		expr: expr,
		args: args,
	}
}
func (we *whereExpr) And(expr string, args ...interface{}) *whereExpr {
	we.expr += " AND " + expr
	we.args = append(we.args, args...)
	return we
}
func (we *whereExpr) Or(expr string, args ...interface{}) *whereExpr {
	we.expr += " OR " + expr
	we.args = append(we.args, args...)
	return we
}
func (we *whereExpr) First(entity interface{}) error {

	return we.db.firstWithFilter(entity, we.expr, we.args...)
}

type onCreateEntity = func(db *TenantDB, entity interface{}) error

var OnCreateEntity onCreateEntity

func (db *TenantDB) Create(entity interface{}) error {
	return OnCreateEntity(db, entity)
}

type DeleteResult struct {
	RowsAffected int64
	Error        error
}

type onDeleteEntity = func(db *TenantDB, entity interface{}, filter string, args ...interface{}) (int64, error)

var OnDeleteEntity onDeleteEntity

func (we *whereExpr) Delete(entity interface{}) DeleteResult {
	rowsAffected, err := OnDeleteEntity(we.db, entity, we.expr, we.args...)
	return DeleteResult{
		RowsAffected: rowsAffected,
		Error:        err,
	}

}
