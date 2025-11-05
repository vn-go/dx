package dx

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	sqlDB "database/sql"

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
func (q *queryObject) ToArray() (any, error) {
	sql, err := q.Analize()
	if err != nil {
		return nil, err
	}
	if Options.ShowSql {
		fmt.Println("-------------------")
		fmt.Println(sql.Query)
		fmt.Println("-------------------")
	}
	rows, err := q.db.Query(sql.Query, sql.Args...)
	if err != nil {
		return nil, err
	}
	return q.db.ScanRowsToArrayStruct(rows, sql.OutputFields.ToStruct(sql.Hash256AccessScope))
}
func (db *DB) ScanRowsToArrayStruct(rows *sqlDB.Rows, returnType reflect.Type) (any, error) {
	defer rows.Close()
	sliceType := reflect.SliceOf(reflect.PointerTo(returnType))
	sliceValue := reflect.MakeSlice(sliceType, 0, 8)
	for rows.Next() {
		item, err := db.ScanRowToStruct(rows, returnType) // dùng hàm ánh xạ struct tối ưu của bạn
		if err != nil {
			return nil, err
		}
		if item != nil {
			sliceValue = reflect.Append(sliceValue, reflect.ValueOf(item))
		}

	}

	// Trả về slice interface{}
	return sliceValue.Interface(), nil
}
func (db *DB) ScanRowToStruct(rows *sqlDB.Rows, returnType reflect.Type) (any, error) {
	dest := reflect.New(returnType).Interface()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	v := reflect.ValueOf(dest).Elem()
	fields := make([]any, len(cols))
	for i, col := range cols {
		f := v.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, col)
		})
		if f.IsValid() && f.CanSet() {
			fields[i] = f.Addr().Interface()
		} else {
			var dummy any
			fields[i] = &dummy
		}
	}

	err = rows.Scan(fields...)
	return dest, err
}

func (db *DB) FindFirst(fromModel any, selector, conditional string, args ...any) (any, error) {
	typ := reflect.TypeOf(fromModel)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if selector == "" {
		selector = typ.Name() + "()"
	} else {
		selector = fmt.Sprintf("%s(%s)", typ.Name(), selector)
	}
	if conditional != "" {
		selector += ",where(" + conditional + ")"
	}
	sql, err := db.Smart(selector, args...)
	if err != nil {
		return nil, err
	}
	returnType := sql.OutputFields.ToStruct(sql.Hash256AccessScope)
	if Options.ShowSql {
		fmt.Println("-------------------")
		fmt.Println(sql.Query)
		fmt.Println("-------------------")
	}
	r, err := db.Query(sql.Query, sql.Args...)
	if err != nil {
		return nil, err
	}

	return db.ScanRowToStruct(r, returnType)
}
func (db *DB) Find(fromModel any, selector, conditional string, args ...any) (any, error) {
	typ := reflect.TypeOf(fromModel)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if selector == "" {
		selector = typ.Name() + "()"
	} else {
		selector = fmt.Sprintf("%s(%s)", typ.Name(), selector)
	}
	if conditional != "" {
		selector += ",where(" + conditional + ")"
	}

	sql, err := db.Smart(selector, args...)
	if err != nil {
		return nil, err
	}

	returnType := sql.OutputFields.ToStruct(sql.Hash256AccessScope)

	rows, err := db.Query(sql.Query, sql.Args...)
	if err != nil {
		return nil, err
	}
	return db.ScanRowsToArrayStruct(rows, returnType)

}

type dataSet struct {
	source        string
	sourceArgs    []any
	filter        string
	filterArgs    []any
	orders        []string
	orderArgs     []any
	limit         *uint64
	offset        *uint64
	selectors     []string
	selectorsArgs []any
	ctx           context.Context
	db            *DB
}

func (db *DB) Dataset() *dataSet {
	return &dataSet{
		db: db,
	}
}
func (db *DB) DatasetWithContext(ctx context.Context) *dataSet {
	return &dataSet{
		db:  db,
		ctx: ctx,
	}
}
func (ds *dataSet) From(source string, args ...any) *dataSet {
	ds.source = source
	ds.sourceArgs = args
	return ds
}
func (ds *dataSet) Where(filter string, args ...any) *dataSet {
	ds.filter = filter
	ds.filterArgs = args
	return ds
}
func (ds *dataSet) Sort(orders ...string) *dataSet {
	ds.orders = orders
	return ds
}
func (ds *dataSet) SortDesc(orders ...string) *dataSet {
	for _, order := range orders {
		ds.orders = append(ds.orders, order+" desc")
	}
	return ds
}
func (ds *dataSet) Limit(limit uint64) *dataSet {
	ds.limit = &limit
	return ds
}
func (ds *dataSet) Offset(offset uint64) *dataSet {
	ds.offset = &offset
	return ds
}
func (ds *dataSet) Select(fields string, args ...any) *dataSet {
	ds.selectors = append(ds.selectors, fields)
	ds.selectorsArgs = append(ds.selectorsArgs, args...)
	return ds
}
func (ds *dataSet) Analize() (*sql.SmartSqlParser, error) {
	args := []any{}
	dslItems := []string{
		fmt.Sprintf("from(%s)", ds.source),
	}
	args = append(args, ds.sourceArgs...)
	if len(ds.selectors) > 0 {
		dslItems = append(dslItems, strings.Join(ds.selectors, ","))
		args = append(args, ds.selectorsArgs...)
	}

	var strFilter string
	if ds.filter != "" {
		strFilter = fmt.Sprintf("where(%s)", ds.filter)
		dslItems = append(dslItems, strFilter)
		args = append(args, ds.filterArgs...)
	}
	var strOrders string
	if len(ds.orders) > 0 {
		strOrders = fmt.Sprintf("sort(%s)", strings.Join(ds.orders, ","))
		dslItems = append(dslItems, strOrders)
		args = append(args, ds.orderArgs...)
	}
	var strLimit string
	if ds.limit != nil {
		strLimit = "take(?)"
		dslItems = append(dslItems, strLimit)
		args = append(args, *ds.limit)
	}
	var strOffset string
	if ds.offset != nil {
		strOffset = "skip(?)"
		dslItems = append(dslItems, strOffset)
		args = append(args, *ds.offset)
	}

	dsl := strings.Join(dslItems, ",")
	return ds.db.Smart(dsl, args...)

}
func (ds *dataSet) ToArray() (any, error) {
	sql, err := ds.Analize()
	if err != nil {
		return nil, err
	}
	if Options.ShowSql {
		fmt.Println("-------------------")
		fmt.Println(sql.Query)
		fmt.Println("-------------------")
	}
	rows, err := ds.db.Query(sql.Query, sql.Args...)
	if err != nil {
		return nil, err
	}
	return ds.db.ScanRowsToArrayStruct(rows, sql.OutputFields.ToStruct(sql.Hash256AccessScope))
}
func (ds *dataSet) First() (any, error) {
	sql, err := ds.Analize()
	if err != nil {
		return nil, err
	}
	if Options.ShowSql {
		fmt.Println("-------------------")
		fmt.Println(sql.Query)
		fmt.Println("-------------------")
	}
	rows, err := ds.db.Query(sql.Query, sql.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		return ds.db.ScanRowToStruct(rows, sql.OutputFields.ToStruct(sql.Hash256AccessScope))
	}
	return nil, nil
}
