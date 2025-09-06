package dx

import (
	"fmt"
	"reflect"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

type modelTypeWhere struct {
	*modelType
	err       error
	whereExpr *whereTypesItem
	lastWhere *whereTypesItem
	limit     *uint64
	offset    *uint64
}

func (m *modelType) Where(args ...interface{}) *modelTypeWhere {
	if len(args) == 0 {
		return &modelTypeWhere{
			err: fmt.Errorf("(db *DB) Where(<requires at least one argument to be passed>,[list of arguments])"),
		}
	}
	if reflect.TypeOf(args[0]) != reflect.TypeFor[string]() {
		return &modelTypeWhere{
			err: fmt.Errorf("(db *DB) Where(<argument must be string>),[list of arguments])"),
		}
	}

	ret := &modelTypeWhere{
		modelType: m,
		whereExpr: &whereTypesItem{
			filter: args[0].(string),
			args:   args[1:],
		},
	}
	ret.lastWhere = ret.whereExpr
	return ret
}
func (w *modelTypeWhere) And(args ...interface{}) *modelTypeWhere {
	if w.err != nil {
		return w
	}
	if len(args) == 0 {
		w.err = fmt.Errorf("(db *DB) Where(<requires at least one argument to be passed>,[list of arguments])")
		return w
	}
	if reflect.TypeOf(args[0]) != reflect.TypeFor[string]() {
		return &modelTypeWhere{
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
func (w *modelTypeWhere) Or(args ...interface{}) *modelTypeWhere {
	if w.err != nil {
		return w
	}
	if len(args) == 0 {
		w.err = fmt.Errorf("(db *DB) Where(<requires at least one argument to be passed>,[list of arguments])")
		return w
	}
	if reflect.TypeOf(args[0]) != reflect.TypeFor[string]() {
		return &modelTypeWhere{
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
func (w *modelTypeWhere) getFilter() (string, []any) {
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
func (m *modelTypeWhere) Count(ret *uint64) error {
	if m.err != nil {
		return m.err
	}
	wherStr, args := m.getFilter()
	key := m.typEle.String() + "//modelTypeWhere/Count" + "/" + wherStr
	query, err := internal.OnceCall(key, func() (string, error) {

		ent, err := model.ModelRegister.GetModelByType(m.typEle)
		if err != nil {
			return "", err
		}
		compiler, err := expr.CompileJoin(ent.Entity.TableName, m.db.DB)
		if err != nil {
			return "", err
		}
		//compiler.Context.Tables = append(compiler.Context.Tables, ent.Entity.TableName)
		compiler.Context.Purpose = expr.BUILD_WHERE
		err = compiler.BuildWhere(wherStr)
		if err != nil {
			return "", err
		}
		dialect := factory.DialectFactory.Create(m.db.DriverName)
		retSql := "select count(*) from " + dialect.Quote(ent.Entity.TableName) + " " + dialect.Quote(compiler.Context.Alias[ent.Entity.TableName]) + " where " + compiler.Content
		return retSql, nil
	})
	if err != nil {
		return err
	}

	// Thực thi câu lệnh SQL và quét kết quả vào biến 'count'
	err = m.db.QueryRow(query, args...).Scan(&ret)
	if err != nil {
		return m.db.parseError(err)
	}
	return nil
}
func (m *modelTypeWhere) Limit(num uint64) *modelTypeWhere {
	m.limit = &num
	return m
}
func (m *modelTypeWhere) Offset(num uint64) *modelTypeWhere {
	m.offset = &num
	return m
}
func (m *modelType) Limit(num uint64) *modelTypeWhere {
	ret := &modelTypeWhere{
		modelType: m,
		limit:     &num,
	}
	ret.lastWhere = ret.whereExpr
	return ret
}
func (m *modelType) Offset(num uint64) *modelTypeWhere {
	ret := &modelTypeWhere{
		modelType: m,
		offset:    &num,
	}
	ret.lastWhere = ret.whereExpr
	return ret
}

//db.Limit(pageSize).Offset(offset).Find(&users)
//SELECT * FROM [users] ORDER BY (SELECT NULL) OFFSET @offset ROWS FETCH NEXT @limit ROWS ONLY;
