package dx

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/sql"
)

type dynamicQuery struct {
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

// Tạo pool toàn cục để tái sử dụng đối tượng dynamicQuery
var dynamicQueryPool = sync.Pool{
	New: func() interface{} {
		return &dynamicQuery{}
	},
}

// Lấy một đối tượng từ pool
func acquireDynamicQuery() *dynamicQuery {
	q := dynamicQueryPool.Get().(*dynamicQuery)

	// Reset toàn bộ fields (đề phòng bị reuse)
	q.selector = ""
	q.selectorArgs = q.selectorArgs[:0]
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

// Trả đối tượng về pool
func releaseDynamicQuery(q *dynamicQuery) {
	// Xoá hết dữ liệu để tránh memory leak hoặc giữ reference
	q.selector = ""
	q.selectorArgs = q.selectorArgs[:0]
	q.source = ""
	q.sourceArgs = q.sourceArgs[:0]
	q.skip = nil
	q.take = nil
	q.orderBy = ""
	q.orderByArgs = q.orderByArgs[:0]
	q.filter = ""
	q.filterArgs = q.filterArgs[:0]

	dynamicQueryPool.Put(q)
}
func (q *dynamicQuery) Join(source string, args ...interface{}) *dynamicQuery {
	q.source = source
	q.sourceArgs = args
	return q
}
func (q *dynamicQuery) Filter(filter string, args ...interface{}) *dynamicQuery {
	if q.filter == "" {
		q.filter = filter
		q.filterArgs = args
	} else {
		q.filter = fmt.Sprintf("%s and %s", q.filter, filter)
		q.filterArgs = append(q.filterArgs, args...)
	}
	return q
}
func (q *dynamicQuery) Skip(skip int) *dynamicQuery {
	q.skip = &skip
	return q
}
func (q *dynamicQuery) Take(take int) *dynamicQuery {
	q.take = &take
	return q
}
func (q *dynamicQuery) OrderBy(orderBy string, args ...interface{}) *dynamicQuery {
	if q.orderBy == "" {
		q.orderBy = orderBy
		q.orderByArgs = args
	} else {
		q.orderBy = fmt.Sprintf("%s, %s", q.orderBy, orderBy)
		q.orderByArgs = append(q.orderByArgs, args...)
	}

	return q
}
func (q *dynamicQuery) Build(db *DB) (query *sql.SmartSqlParser, err error) {
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
func (q *dynamicQuery) buildForGetFirst(db *DB) (query *sql.SmartSqlParser, err error) {
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

	str = append(str, " take(?)")
	args = append(args, 1)
	query, err = db.Smart(strings.Join(str, " ,"), args...)
	return query, err
}

func (q *dynamicQuery) ToArray(db *DB) (any, error) {
	defer releaseDynamicQuery(q)
	// Build query
	query, err := q.Build(db)
	if err != nil {
		return nil, err
	}
	if Options.ShowSql {
		fmt.Println("----------------------------------")
		fmt.Println(query.Query)
		fmt.Println("----------------------------------")
	}
	rows, err := db.Query(query.Query, query.Args...)
	if err != nil {
		return nil, err
	}
	returnType := query.OutputFields.ToArrayOfStruct(query.OutputFields.ToHas256Key())
	ret, err := db.ScanRowsToArrayStruct(rows, returnType)
	return ret, err
}
func (q *dynamicQuery) ToArrayWithContext(ctx context.Context, db *DB) (any, error) {
	defer releaseDynamicQuery(q)
	// Build query
	query, err := q.Build(db)
	if err != nil {
		return nil, err
	}
	if Options.ShowSql {
		fmt.Println("----------------------------------")
		fmt.Println(query.Query)
		fmt.Println("----------------------------------")
	}
	rows, err := db.QueryContext(ctx, query.Query, query.Args...)
	if err != nil {
		return nil, err
	}
	returnType := query.OutputFields.ToStruct(query.OutputFields.ToHas256Key())
	ret, err := db.ScanRowsToArrayStruct(rows, returnType)
	return ret, err
}
func (q *dynamicQuery) ToItem(db *DB) (any, error) {
	defer releaseDynamicQuery(q)
	// Build query
	query, err := q.buildForGetFirst(db)
	if err != nil {
		return nil, err
	}
	if Options.ShowSql {
		fmt.Println("----------------------------------")
		fmt.Println(query.Query)
		fmt.Println("----------------------------------")
	}
	rows, err := db.Query(query.Query, query.Args...)
	if err != nil {
		return nil, err
	}
	returnType := query.OutputFields.ToStruct(query.OutputFields.ToHas256Key())
	return db.ScanRowToStruct(rows, returnType)
}
func (q *dynamicQuery) ToItemWithContext(ctx context.Context, db *DB) (any, error) {
	defer releaseDynamicQuery(q)
	// Build query
	query, err := q.buildForGetFirst(db)
	if err != nil {
		return nil, err
	}
	if Options.ShowSql {
		fmt.Println("----------------------------------")
		fmt.Println(query.Query)
		fmt.Println("----------------------------------")
	}
	rows, err := db.QueryContext(ctx, query.Query, query.Args...)
	if err != nil {
		return nil, err
	}
	returnType := query.OutputFields.ToStruct(query.OutputFields.ToHas256Key())
	return db.ScanRowToStruct(rows, returnType)

}

func NewDynamicQuery(selector string, args ...interface{}) *dynamicQuery {
	return acquireDynamicQuery()
}
