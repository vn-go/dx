package dx

import (
	"context"
	"sync"

	"github.com/vn-go/dx/sql"
)

// dx.enduser.query.frontEndQuery.go
type frontEndQuery struct {
	sqlInfo       *sql.ExtractInfo
	db            *DB
	err           error
	selector      string
	selectorsArgs []any

	selectorsField    []frontEndQueryResult
	selectorsFieldMap map[string]frontEndQueryResult
	filter            string
	filterArgs        []any

	filterField []frontEndQueryResult
	args        sql.Args
	OutptFields sql.ExtractInfoOutputField
}

func (f *frontEndQuery) ToDyanmicArrayWithContext(ctx context.Context) (any, error) {
	defer putFrontEndQuery(f)
	queryCompiler, err := f.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := f.db.QueryContext(ctx, queryCompiler.Query, queryCompiler.Args...)
	if err != nil {
		return nil, err
	}
	return f.db.ScanRowsToArrayStruct(rows, queryCompiler.Output.OutputFields.ToStruct(queryCompiler.Output.Hash256))
}

func (f *frontEndQuery) ToDyanmicArray() (any, error) {
	defer putFrontEndQuery(f)
	queryCompiler, err := f.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := f.db.Query(queryCompiler.Query, queryCompiler.Args...)
	if err != nil {
		return nil, err
	}
	return f.db.ScanRowsToArrayStruct(rows, queryCompiler.Output.OutputFields.ToStruct(queryCompiler.Output.Hash256))
}

func (f *frontEndQuery) Filter(s string, args ...any) *frontEndQuery {
	if f.filter == "" {
		f.filter = s
		f.filterArgs = args
	} else {
		f.filter += " AND " + s
		f.filterArgs = append(f.filterArgs, args...)
	}
	return f
}
func (f *frontEndQuery) FilterOr(s string, args ...any) *frontEndQuery {
	if f.filter == "" {
		f.filter = s
		f.filterArgs = args
	} else {
		f.filter += " OR " + s
		f.filterArgs = append(f.filterArgs, args...)
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
	q.filterArgs = q.filterArgs[:0]
	q.selectorsArgs = q.selectorsArgs[:0]
	q.args = q.args[:0]
	q.OutptFields.Hash256 = ""
	q.OutptFields.OutputFieldMap = nil
	q.OutptFields.OutputFields = nil

	q.db = nil
	q.err = nil
	frontEndQueryPool.Put(q)
}
func (e *endUserQuery) ToFrontEnd(db *DB) *frontEndQuery {
	ret := frontEndQueryPool.Get().(*frontEndQuery)
	sqlInfo, _, err := e.ToSql(db)
	if err != nil {
		ret.err = err
		return ret
	}

	// if err != nil {
	// 	ret.err = err
	// 	return ret
	// }
	ret.db = db
	ret.sqlInfo = sqlInfo.Clone()
	// ret.args = args
	ret.filter = e.filter
	ret.filterArgs = e.filterArgs
	ret.selector = e.selector
	ret.selectorsArgs = e.selectorArgs
	ret.args = sqlInfo.Args
	ret.selectorsFieldMap = make(map[string]frontEndQueryResult)
	ret.OutptFields = sql.ExtractInfoOutputField{
		OutputFields:   sqlInfo.OuputInfo.OutputFields,
		OutputFieldMap: sqlInfo.OuputInfo.OutputFieldMap,
		Hash256:        sqlInfo.OuputInfo.Hash256,
	}

	return ret
}
