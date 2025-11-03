package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/sql"
)

type queryObject struct {
	db            *DB
	source        []string
	sourceArgs    []any
	filter        string
	filterArgs    []any
	orders        []string
	orderArgs     []any
	limit         *uint64
	offset        *uint64
	selectors     []string
	selectorsArgs []any
	tables        []string
}

func (q *queryObject) Count() (int64, error) {
	args := []any{}
	dslItems := []string{
		fmt.Sprintf("from(%s)", strings.Join(q.source, ",")),
	}
	args = append(args, q.sourceArgs...)
	var strFilter string
	if q.filter != "" {
		strFilter = fmt.Sprintf("where(%s)", q.filter)
		dslItems = append(dslItems, strFilter)
		args = append(args, q.filterArgs...)
	}

	dslItems = append(dslItems, "count(*) ItemCount")

	dsl := strings.Join(dslItems, ",")
	sql, err := q.db.Smart(dsl, args...)
	if err != nil {
		return 0, err
	}
	if Options.ShowSql {
		fmt.Println("-------------------")
		fmt.Println(sql.Query)
		fmt.Println("-------------------")
	}
	r, err := q.db.Query(sql.Query, sql.Args...)
	if err != nil {
		return 0, err
	}
	var ret int64
	defer r.Close()
	for r.Next() {
		err = r.Scan(&ret)
		return ret, err
	}
	return 0, nil

}

func QueryItems[TResult any](db *DB, dsl string, args ...any) ([]TResult, error) {
	var items []TResult
	err := db.DslQuery(&items, dsl, args...)
	if err != nil {
		return nil, err
	}
	return items, nil
}
func (db *DB) QueryModel(model any) *queryObject {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return &queryObject{
		db:            db,
		source:        []string{typ.Name()},
		tables:        []string{typ.Name()},
		sourceArgs:    []any{},
		filterArgs:    []any{},
		orderArgs:     []any{},
		selectorsArgs: []any{},
	}
}
func (q *queryObject) LeftJoin(model any, on string, args ...any) *queryObject {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	q.source = append(q.source, typ.Name())
	q.source = append(q.source, "left("+on+")")
	q.sourceArgs = append(q.sourceArgs, args...)
	q.tables = append(q.tables, typ.Name())

	return q
}
func (q *queryObject) RightJoin(model any, on string, args ...any) *queryObject {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	q.source = append(q.source, typ.Name())
	q.source = append(q.source, "right("+on+")")
	q.sourceArgs = append(q.sourceArgs, args...)
	q.tables = append(q.tables, typ.Name())
	return q
}
func (q *queryObject) InnerJoin(model any, on string, args ...any) *queryObject {
	typ := reflect.TypeOf(model)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	q.source = append(q.source, typ.Name())
	q.source = append(q.source, on)
	q.sourceArgs = append(q.sourceArgs, args...)
	q.tables = append(q.tables, typ.Name())
	return q
}
func (q *queryObject) And(filter string, args ...any) *queryObject {
	if q.filter == "" {
		q.filter = filter
	} else {
		q.filter = "(" + q.filter + ") and (" + filter + ")"
	}
	q.filterArgs = append(q.filterArgs, args...)
	return q
}

func (q *queryObject) Or(filter string, args ...any) *queryObject {
	if q.filter == "" {
		q.filter = filter
	} else {
		q.filter = "(" + q.filter + ") or (" + filter + ")"
	}
	q.filterArgs = append(q.filterArgs, args...)
	return q
}
func (q *queryObject) Sort(orders ...string) *queryObject {
	for _, order := range orders {
		q.orders = append(q.orders, strings.Split(order, ",")...)
	}
	return q
}
func (q *queryObject) SortDesc(orders ...string) *queryObject {
	for _, order := range orders {
		for _, o := range strings.Split(order, ",") {
			q.orders = append(q.orders, o+" desc")
		}
	}
	return q
}
func (q *queryObject) Limit(limit uint64) *queryObject {
	q.limit = &limit
	return q
}
func (q *queryObject) Offset(offset uint64) *queryObject {
	q.offset = &offset
	return q
}
func (q *queryObject) Select(fields string, args ...any) *queryObject {
	q.selectors = append(q.selectors, fields)
	q.selectorsArgs = append(q.selectorsArgs, args...)
	return q
}
func (q *queryObject) Analize() (*sql.SmartSqlParser, error) {
	args := []any{}
	dslItems := []string{
		fmt.Sprintf("from(%s)", strings.Join(q.source, ",")),
	}
	args = append(args, q.sourceArgs...)
	if len(q.selectors) > 0 {
		dslItems = append(dslItems, strings.Join(q.selectors, ","))
		args = append(args, q.selectorsArgs...)
	}

	var strFilter string
	if q.filter != "" {
		strFilter = fmt.Sprintf("where(%s)", q.filter)
		dslItems = append(dslItems, strFilter)
		args = append(args, q.filterArgs...)
	}
	var strOrders string
	if len(q.orders) > 0 {
		strOrders = fmt.Sprintf("sort(%s)", strings.Join(q.orders, ","))
		dslItems = append(dslItems, strOrders)
		args = append(args, q.orderArgs...)
	}
	var strLimit string
	if q.limit != nil {
		strLimit = "take(?)"
		dslItems = append(dslItems, strLimit)
		args = append(args, *q.limit)
	}
	var strOffset string
	if q.offset != nil {
		strOffset = "skip(?)"
		dslItems = append(dslItems, strOffset)
		args = append(args, *q.offset)
	}

	dsl := strings.Join(dslItems, ",")
	return q.db.Smart(dsl, args...)

}
