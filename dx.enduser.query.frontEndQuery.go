package dx

import (
	"sync"

	"github.com/vn-go/dx/sql"
)

// dx.enduser.query.frontEndQuery.go
type frontEndQuery struct {
	args              []any
	sqlInfo           *sql.ExtractInfo
	db                *DB
	err               error
	selector          string
	selectorsArgs     []any
	selectorsField    []frontEndQueryResult
	selectorsFieldMap map[string]frontEndQueryResult
	filter            string
	filterArgs        []any
	filterField       []frontEndQueryResult
}

func (f *frontEndQuery) Filter(s string, args ...any) *frontEndQuery {
	if f.filter == "" {
		f.filter = s
		f.filterArgs = args
	} else {
		f.filter += " AND " + s
		f.args = append(f.args, args...)
	}
	return f
}
func (f *frontEndQuery) FilterOr(s string, args ...any) *frontEndQuery {
	if f.filter == "" {
		f.filter = s
		f.filterArgs = args
	} else {
		f.filter += " OR " + s
		f.args = append(f.args, args...)
	}
	return f
}
func (f *frontEndQuery) Select(selector string, args ...any) *frontEndQuery {
	if f.err != nil {
		return f

	}
	f.selector = selector
	f.selectorsArgs = args
	return f
}

var frontEndQueryPool = sync.Pool{
	New: func() any {
		return new(frontEndQuery)
	},
}

func putFrontEndQuery(q *frontEndQuery) {
	if q == nil {
		return
	}
	q.args = q.args[:0]
	q.db = nil
	q.err = nil
	frontEndQueryPool.Put(q)
}
func (e *endUserQuery) ToFrontEnd(db *DB) *frontEndQuery {
	ret := frontEndQueryPool.Get().(*frontEndQuery)
	sqlInfo, args, err := e.ToSql(db)
	if err != nil {
		ret.err = err
		return ret
	}
	ret.args = args
	// if err != nil {
	// 	ret.err = err
	// 	return ret
	// }
	ret.db = db
	ret.sqlInfo = sqlInfo.Clone()
	return ret
}
