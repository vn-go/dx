package dx

import (
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
}
type findResult struct {
	RowsAffected uint64
	Error        error
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
	return w.db.firstWithFilter(item, whereStr, ars...)
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
	return w.db.findtWithFilter(item, whereStr, orderStr, ars...)
}
