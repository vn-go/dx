package dx

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/sql"
)

type endUserQuery struct {
	selector     string
	selectorArgs []any
	source       string
	sourceArgs   []any
	filter       string
	filterArgs   []any
	order        string
	orderArgs    []any
	limit        *int
	offset       *int
	dslQuery     string
	dslQueryArgs []any
}

var endUserQueryPool = sync.Pool{
	New: func() any {
		return new(endUserQuery)
	},
}

func putEndUserQuery(q *endUserQuery) {
	if q == nil {
		return
	}

	// reset toàn bộ field để tránh leak data cũ
	q.selector = ""
	q.selectorArgs = q.selectorArgs[:0]
	q.source = ""
	q.sourceArgs = q.sourceArgs[:0]
	q.filter = ""
	q.filterArgs = q.filterArgs[:0]
	q.order = ""
	q.orderArgs = q.orderArgs[:0]
	q.limit = nil
	q.offset = nil
	q.dslQuery = ""
	q.dslQueryArgs = q.dslQueryArgs[:0]

	endUserQueryPool.Put(q)
}
func (e *endUserQuery) IsCompileError(err error) *sql.CompilerError {
	if compilerErr, ok := err.(*sql.CompilerError); ok {
		return compilerErr
	}
	return nil
}
func (e *endUserQuery) ToSql(db *DB) (*sql.ExtractInfo, []any, error) {

	if e.dslQuery == "" {
		dlsQuery, args := e.toDlsQuery()

		ret, err := sql.Compiler.ExtractInfo(db.Dialect, dlsQuery, args)

		if err != nil {
			return nil, nil, err
		}
		retAgrs, err := ret.Args.ToArray(args)
		if err != nil {
			return nil, nil, err
		}
		return ret, retAgrs, nil

	} else {
		ret, err := sql.Compiler.ExtractInfo(db.Dialect, e.dslQuery, e.dslQueryArgs)
		if err != nil {
			return nil, nil, err
		}
		retAgrs, err := ret.Args.ToArray(e.dslQueryArgs)
		if err != nil {
			return nil, nil, err
		}
		return ret, retAgrs, nil

	}

}
func (e *endUserQuery) ToArray(db *DB) (any, error) {
	sqlResult, args, err := e.ToSql(db)
	if err != nil {
		return nil, err
	}
	query := db.Dialect.GetSelectStatement(sqlResult.SelectStatement)
	retArgs, err := sqlResult.Args.ToArray(args)
	if err != nil {
		return nil, err
	}
	returnType := sqlResult.OuputInfo.OutputFields.ToStruct(sqlResult.OuputInfo.Hash256)

	return db.ExecToArrayByType(returnType, query, retArgs...)

}
func (e *endUserQuery) toDlsQuery() (string, []any) {
	str := []string{}
	args := []any{}
	if e.selector != "" {
		str = append(str, e.selector)
		args = append(args, e.selectorArgs...)
	}
	if e.source != "" {
		str = append(str, fmt.Sprintf("from (%s)", e.source))
		args = append(args, e.sourceArgs...)
	}
	if e.filter != "" {
		str = append(str, fmt.Sprintf("where (%s)", e.filter))
		args = append(args, e.filterArgs...)
	}
	if e.order != "" {
		str = append(str, fmt.Sprintf("sort(%s)", e.order))
		args = append(args, e.orderArgs...)
	}

	return strings.Join(str, ", "), args
}

func (e *endUserQuery) Filter(filter string, args ...any) *endUserQuery {
	if e.filter == "" {
		e.filter = filter
		e.filterArgs = args
	} else {
		e.filter += " and " + filter
		e.filterArgs = append(e.filterArgs, args...)
	}
	return e
}

func (e *endUserQuery) Source(source string, args ...any) *endUserQuery {
	e.source = source
	e.sourceArgs = args
	return e
}

func (e *endUserQuery) Fields(selector string, args ...any) *endUserQuery {
	if e.selector == "" {
		e.selector = selector
		e.selectorArgs = args
	} else {
		e.selector += ", " + selector
		e.selectorArgs = append(e.selectorArgs, args...)
	}
	return e
}
func (e *endUserQuery) Query(source string, args ...any) *endUserQuery {
	e.dslQuery = source
	e.dslQueryArgs = args
	return e
}
func NewEndUserQuery() *endUserQuery {
	return endUserQueryPool.Get().(*endUserQuery)

}
