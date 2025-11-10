package dx

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sql"
)

// -------------------
type dslQuery[TResult any] struct {
	selector     string
	selectorArgs []interface{}
	source       string
	sourceArgs   []interface{}
	skip         *int
	take         *int
	orderBy      string
	orderByArgs  []interface{}
	filter       string
	filterArgs   []interface{}
}

// reset: trả về trạng thái "mới" trước khi Put vào pool
func (q *dslQuery[TResult]) reset() {
	q.selector = ""
	if q.selectorArgs != nil {
		// giữ slice but set len=0 để reuse backing array
		q.selectorArgs = q.selectorArgs[:0]
	}
	q.source = ""
	if q.sourceArgs != nil {
		q.sourceArgs = q.sourceArgs[:0]
	}
	q.skip = nil
	q.take = nil
	q.orderBy = ""
	if q.orderByArgs != nil {
		q.orderByArgs = q.orderByArgs[:0]
	}
	q.filter = ""
	if q.filterArgs != nil {
		q.filterArgs = q.filterArgs[:0]
	}
}

type initNewDSLQueryPool struct {
	once sync.Once
	pool *sync.Pool
}

var initNewDSLQueryPoolSyncMap sync.Map

// newDSLQueryPool returns a singleton *sync.Pool per TResult type.
func newOrGetDSLQueryPool[TResult any]() *sync.Pool {
	key := reflect.TypeFor[TResult]() // or reflect.TypeOf((*TResult)(nil)).Elem()
	if key.Kind() == reflect.Ptr {
		key = key.Elem()
	}

	actual, _ := initNewDSLQueryPoolSyncMap.LoadOrStore(key, &initNewDSLQueryPool{})
	entry := actual.(*initNewDSLQueryPool)

	entry.once.Do(func() {
		entry.pool = &sync.Pool{
			New: func() interface{} {
				return &dslQuery[TResult]{}
			},
		}
	})

	return entry.pool
}

// ReleaseDSLQuery: reset rồi trả object vào pool
func releaseDSLQuery[TResult any](p *sync.Pool, q *dslQuery[TResult]) {
	if q == nil {
		return
	}
	q.reset()
	p.Put(q)
}
func (q *dslQuery[TResult]) Join(source string, args ...interface{}) *dslQuery[TResult] {
	q.source = source
	q.sourceArgs = args
	return q
}
func (q *dslQuery[TResult]) Filter(filter string, args ...interface{}) *dslQuery[TResult] {
	if q.filter == "" {
		q.filter = filter
		q.filterArgs = args
	} else {
		q.filter = fmt.Sprintf("%s and %s", q.filter, filter)
		q.filterArgs = append(q.filterArgs, args...)
	}
	return q
}
func (q *dslQuery[TResult]) Skip(skip int) *dslQuery[TResult] {
	q.skip = &skip
	return q
}
func (q *dslQuery[TResult]) Take(take int) *dslQuery[TResult] {
	q.take = &take
	return q
}
func (q *dslQuery[TResult]) OrderBy(orderBy string, args ...interface{}) *dslQuery[TResult] {
	if q.orderBy == "" {
		q.orderBy = orderBy
		q.orderByArgs = args
	} else {
		q.orderBy = fmt.Sprintf("%s, %s", q.orderBy, orderBy)
		q.orderByArgs = append(q.orderByArgs, args...)
	}

	return q
}

type initgetColumnsName struct {
	cols   []string
	mapCol map[string]fieldInfo
	sync.Once
}

var initgetColumnsNamCache sync.Map

func (q *dslQuery[TResult]) getColumnsName() ([]string, map[string]fieldInfo) {
	key := reflect.TypeFor[TResult]()
	if key.Kind() == reflect.Ptr {
		key = key.Elem()
	}
	a, _ := initgetColumnsNamCache.LoadOrStore(key, &initgetColumnsName{})
	i := a.(*initgetColumnsName)
	i.Do(func() {
		ret := []string{}
		retMap := map[string]fieldInfo{}
		t := reflect.TypeFor[TResult]()
		for i := 0; i < t.NumField(); i++ {
			ret = append(ret, t.Field(i).Name)
			retMap[t.Field(i).Name] = fieldInfo{
				offset: t.Field(i).Offset,
				typ:    t.Field(i).Type,
			}

		}
		i.cols = ret
		i.mapCol = retMap
	})
	return i.cols, i.mapCol

}
func (q *dslQuery[TResult]) Build(db *DB) (query *sql.SmartSqlParser, err error) {
	str := []string{q.selector}
	args := append([]interface{}{}, q.selectorArgs...)

	if q.source != "" {
		str = append(str, " from("+q.source+")")
		args = append(args, q.sourceArgs...)
	}
	if q.filter != "" {

		str = append(str, "where("+q.filter+")")
		args = append(args, q.filterArgs...)
	}
	if q.orderBy != "" {

		str = append(str, "sort("+q.orderBy+")")
		args = append(args, q.orderByArgs...)
	}
	if q.skip != nil {
		str = append(str, " skip(?)")
		args = append(args, *q.skip)
	}
	if q.take != nil {
		str = append(str, " take(?)")
		args = append(args, *q.take)
	}
	query, err = db.Smart(strings.Join(str, " ,"), args...)
	return query, err
}
func (q *dslQuery[TResult]) buildForGetFirst(db *DB) (query *sql.SmartSqlParser, err error) {
	str := []string{q.selector}
	args := append([]interface{}{}, q.selectorArgs...)

	if q.source != "" {
		str = append(str, " from("+q.source+")")
		args = append(args, q.sourceArgs...)
	}
	if q.filter != "" {

		str = append(str, "where("+q.filter+")")
		args = append(args, q.filterArgs...)
	}
	if q.orderBy != "" {

		str = append(str, "sort("+q.orderBy+")")
		args = append(args, q.orderByArgs...)
	}

	str = append(str, " take(?)")
	args = append(args, 1)
	query, err = db.Smart(strings.Join(str, " ,"), args...)
	return query, err
}
func (q *dslQuery[TResult]) ToArray(db *DB) ([]TResult, error) {
	// Đảm bảo luôn trả lại object về pool
	defer releaseDSLQuery(newOrGetDSLQueryPool[TResult](), q)

	// Build query
	query, err := q.Build(db)
	if err != nil {
		return nil, err
	}

	// Prepare ị slice of result
	var ret []TResult
	sliceVal := reflect.ValueOf(&ret).Elem()

	// Thực thi query
	rows, err := db.Query(query.Query, query.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Lấy thông tin cột và ánh xạ dữ liệu
	cols, fetchInfo := q.getColumnsName()
	fetchUnsafe(rows, sliceVal.Addr().Interface(), cols, fetchInfo)

	return ret, nil
}
func (q *dslQuery[TResult]) ToItem(db *DB) (*TResult, error) {
	defer releaseDSLQuery(newOrGetDSLQueryPool[TResult](), q)
	query, err := q.buildForGetFirst(db)
	if err != nil {
		return nil, err
	}
	var ret []TResult
	sliceVal := reflect.ValueOf(&ret).Elem()
	if db.DriverName == "mysql" {
		query.Query, query.Args, err = internal.Helper.FixParam(query.Query, query.Args)
		if err != nil {
			return nil, err
		}
	}
	rows, err := db.Query(query.Query, query.Args...)
	if err != nil {
		return nil, err
	}

	_, fectInfo := q.getColumnsName()
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	err = fetchUnsafe(rows, sliceVal.Addr().Interface(), cols, fectInfo)
	if err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, nil
	}
	return &ret[0], nil

}
func (q *dslQuery[TResult]) ToItemWithContext(ctx context.Context, db *DB) (*TResult, error) {
	defer releaseDSLQuery(newOrGetDSLQueryPool[TResult](), q)
	query, err := q.buildForGetFirst(db)
	if err != nil {
		return nil, err
	}
	var ret []TResult
	sliceVal := reflect.ValueOf(&ret).Elem()
	if db.DriverName == "mysql" {
		query.Query, query.Args, err = internal.Helper.FixParam(query.Query, query.Args)
		if err != nil {
			return nil, err
		}
	}
	rows, err := db.QueryContext(ctx, query.Query, query.Args...)
	if err != nil {
		return nil, err
	}

	_, fectInfo := q.getColumnsName()
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	err = fetchUnsafe(rows, sliceVal.Addr().Interface(), cols, fectInfo)
	if err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return nil, nil
	}
	return &ret[0], nil

}
func (q *dslQuery[TResult]) ToArrayWithContext(ctx context.Context, db *DB) ([]TResult, error) {
	defer releaseDSLQuery(newOrGetDSLQueryPool[TResult](), q)
	query, err := q.Build(db)
	if err != nil {
		return nil, err
	}
	var ret []TResult
	sliceVal := reflect.ValueOf(&ret).Elem()
	if db.DriverName == "mysql" {
		query.Query, query.Args, err = internal.Helper.FixParam(query.Query, query.Args)
		if err != nil {
			return nil, err
		}
	}
	rows, err := db.QueryContext(ctx, query.Query, query.Args...)
	if err != nil {
		return nil, err
	}
	cols, fectInfo := q.getColumnsName()
	defer rows.Close()
	fetchUnsafe(rows, sliceVal.Addr().Interface(), cols, fectInfo)
	return ret, nil

}
func NewQuery[TResult any](selector string, args ...interface{}) *dslQuery[TResult] {

	// get correspoding pool for TResult (init if not exist)
	pool := newOrGetDSLQueryPool[TResult]()

	// get object from pool
	q := pool.Get().(*dslQuery[TResult])

	// Re-init (looks like constructor)
	q.selector = selector
	q.selectorArgs = append(q.selectorArgs[:0], args...) // giữ lại backing array để tái sử dụng
	q.source = ""
	q.sourceArgs = q.sourceArgs[:0]
	q.skip = nil
	q.take = nil
	q.orderBy = ""
	q.orderByArgs = q.orderByArgs[:0]
	q.filter = ""
	q.filterArgs = q.filterArgs[:0]

	return q
}
