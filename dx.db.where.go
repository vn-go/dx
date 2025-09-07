package dx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type whereTypesItem struct {
	filter string
	args   []any
	op     string
	next   *whereTypesItem
}
type whereTypes struct {
	db        *DB
	err       error
	whereExpr *whereTypesItem
	lastWhere *whereTypesItem
	orders    []string
	limit     *uint64
	offset    *uint64
	ctx       context.Context
	sqlTx     *sql.Tx
}
type dbContext struct {
	ctx context.Context
	*DB
}

func (c *dbContext) Begin() *Tx {
	sqlTx, err := c.DB.BeginTx(c.ctx, nil)
	if err != nil {
		return &Tx{
			Error: err,
		}
	}
	return sqlTx
}
func (db *DB) WithContext(ctx context.Context) *dbContext {
	return &dbContext{
		ctx: ctx,
		DB:  db,
	}
}
func (tx *Tx) Where(args ...interface{}) *whereTypes {

	ret := tx.db.Where(args...)
	ret.sqlTx = tx.Tx
	return ret
}
func (db *dbContext) Where(args ...interface{}) *whereTypes {
	ret := db.DB.Where(args...)
	ret.ctx = db.ctx
	return ret
}
func (db *DB) Where(args ...interface{}) *whereTypes {
	if len(args) == 0 {
		return &whereTypes{
			err: fmt.Errorf("(db *DB) Where(<requires at least one argument to be passed>,[list of arguments])"),
		}
	}
	if reflect.TypeOf(args[0]) != reflect.TypeFor[string]() {
		return &whereTypes{
			err: fmt.Errorf("(db *DB) Where(<argument must be string>),[list of arguments])"),
		}
	}
	ret := &whereTypes{
		db: db,
		whereExpr: &whereTypesItem{
			filter: args[0].(string),
			args:   args[1:],
		},
		orders: []string{},
	}
	ret.lastWhere = ret.whereExpr
	return ret
}
func (w *whereTypes) And(args ...interface{}) *whereTypes {
	if w.err != nil {
		return w
	}
	if len(args) == 0 {
		w.err = fmt.Errorf("(db *DB) Where(<requires at least one argument to be passed>,[list of arguments])")
		return w
	}
	if reflect.TypeOf(args[0]) != reflect.TypeFor[string]() {
		return &whereTypes{
			err: fmt.Errorf("(db *DB) Where(<argument must be string>),[list of arguments])"),
		}
	}
	w.lastWhere.op = "AND"
	lastWhere := &whereTypesItem{
		filter: args[0].(string),
		args:   args[1:],
	}
	w.lastWhere.next = lastWhere
	w.lastWhere = lastWhere
	// w.args = append(w.args, args[1:]...)
	// w.whereItems = append(w.whereItems, args[0].(string))
	return w
}
func (w *whereTypes) Or(args ...interface{}) *whereTypes {
	if w.err != nil {
		return w
	}
	if len(args) == 0 {
		w.err = fmt.Errorf("(db *DB) Where(<requires at least one argument to be passed>,[list of arguments])")
		return w
	}
	if reflect.TypeOf(args[0]) != reflect.TypeFor[string]() {
		return &whereTypes{
			err: fmt.Errorf("(db *DB) Where(<argument must be string>),[list of arguments])"),
		}
	}
	w.lastWhere.op = "OR"
	lastWhere := &whereTypesItem{
		filter: args[0].(string),
		args:   args[1:],
	}
	w.lastWhere.next = lastWhere
	w.lastWhere = lastWhere
	return w
}
func (w *whereTypes) getFilter() (string, []any) {
	if w.whereExpr == nil {
		return "", nil
	}
	ret := w.whereExpr.filter
	args := w.whereExpr.args
	if w.whereExpr.next != nil {
		op := w.whereExpr.op
		w.whereExpr = w.whereExpr.next
		next, nextArg := w.getFilter()
		ret = ret + " " + op + " " + next
		args = append(args, nextArg...)

	}
	return ret, args

}
func (w *whereTypes) First(item any) error {
	if w.err != nil {
		return w.err
	}
	whereStr, ars := w.getFilter()

	return w.db.firstWithFilter(item, whereStr, w.ctx, w.sqlTx, ars...)
}
func (w *whereTypes) Order(order string) *whereTypes {
	w.orders = append(w.orders, strings.Split(order, ",")...)
	return w
}

func (w *whereTypes) Find(item any) error {
	if w.err != nil {
		return w.err
	}
	whereStr, ars := w.getFilter()
	orderStr := ""
	if len(w.orders) > 0 {
		orderStr = strings.Join(w.orders, ",")
	}
	return w.db.findtWithFilter(item, w.ctx, w.sqlTx, whereStr, orderStr, w.limit, w.offset, true, ars...)
}
func (w *whereTypes) AddTo(item any) error {
	if w.err != nil {
		return w.err
	}
	whereStr, ars := w.getFilter()
	orderStr := ""
	if len(w.orders) > 0 {
		orderStr = strings.Join(w.orders, ",")
	}
	return w.db.findtWithFilter(item, w.ctx, w.sqlTx, whereStr, orderStr, w.limit, w.offset, false, ars...)
}

// for sql server
//
// SELECT * FROM [users] ORDER BY (SELECT NULL) OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY;
func (m *whereTypes) Limit(num uint64) *whereTypes {
	m.limit = &num
	return m
}
func (db *DB) Limit(num uint64) *whereTypes {
	ret := &whereTypes{
		db:     db,
		orders: []string{},
		limit:  &num,
	}
	return ret
}
func (db *DB) Offset(num uint64) *whereTypes {
	ret := &whereTypes{
		db:     db,
		orders: []string{},
		offset: &num,
	}
	return ret
}
func (m *whereTypes) Offset(num uint64) *whereTypes {
	m.offset = &num
	return m
}

//db.Limit(pageSize).Offset(offset).Find(&users)
//SELECT * FROM [users] ORDER BY (SELECT NULL) OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY;
