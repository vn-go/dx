package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
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
	args      internal.SelectorTypesArgs
	strWhere  string
	orders    []string
	strSort   string
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
		args: internal.SelectorTypesArgs{},
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
	m.args.ArgWhere = args
	key := m.typEle.String() + "/modelTypeWhere/Count//modelTypeWhere/Count" + "/" + wherStr

	query, err := internal.OnceCall(key, func() (*types.SqlParse, error) {

		ent, err := model.ModelRegister.GetModelByType(m.typEle)
		if err != nil {
			return nil, err
		}

		retSql := "select count(*) from " + ent.Entity.TableName + " where " + wherStr
		sqlInfo, err := compiler.Compile(retSql, m.db.DriverName, false, false)
		if err != nil {
			return nil, err
		}
		sqlArgs := m.args.GetFields()
		sqlInfo.Info.FieldArs = *sqlArgs
		ret := factory.DialectFactory.Create(m.db.DriverName)
		return ret.BuildSql(sqlInfo.Info)

	})
	if err != nil {
		return err
	}
	argsExec := m.args.GetArgs(query.ArgIndex)
	// exec SQL then scan row to 'count'
	if m.db.DriverName == "mysql" {
		query.Sql, argsExec, err = internal.Helper.FixParam(query.Sql, argsExec)
		if err != nil {
			return err
		}
	}
	err = m.db.QueryRow(query.Sql, argsExec...).Scan(ret)
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
func (m *modelTypeWhere) GetSQL() (string, []any, error) {
	key := m.typEle
	m.strWhere, m.args.ArgWhere = m.getFilter()

	selectSql, err := internal.OnceCall(key, func() (*types.SqlParse, error) {
		var err error
		ent, err := model.ModelRegister.GetModelByType(m.typEle)
		if err != nil {
			return nil, err
		}
		fields := []string{}
		for _, f := range ent.Entity.Cols {
			fields = append(fields, f.Name+" "+f.Field.Name)
		}
		sqlInfo := &types.SqlInfo{
			StrSelect: strings.Join(fields, ","),
			From:      ent.Entity.TableName,
			Limit:     m.limit,
			StrWhere:  m.strWhere,
			Offset:    m.offset,
		}

		sql, err := compiler.GetSql(sqlInfo, m.db.DriverName)
		if err != nil {
			return nil, err
		}

		return sql, nil

	})

	return selectSql.Sql, nil, err
}

// func (s *modelTypeWhere) getKey() string {

// 	s.strWhere, s.args.ArgWhere = s.getFilter()
// 	if s.orders != nil {
// 		s.strSort = strings.Join(s.orders, ",")
// 	}

// 	key := s.strSelect + "+/" + s.strSort + "/" + s.strWhere + "/" + s.strGroup + "/" + s.strHaving + "/" + s.strJoin + "/" + s.strHaving + "/"
// 	if s.limit != nil {
// 		key += "/" + fmt.Sprintf("%d", *s.limit)
// 	}
// 	if s.offset != nil {
// 		key += "/" + fmt.Sprintf("%d", *s.offset)
// 	}

// 	return key

// }
func (m *modelTypeWhere) Find() (any, error) {
	sql, _, err := m.GetSQL()
	if err != nil {
		return nil, err
	}

	ret, err := m.db.fecthItemsOfType(m.typEle, sql, m.ctx, nil, false)
	if err != nil {
		return nil, err
	}

	return ret.Elem().Interface(), err

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

func (m *modelTypeWhere) Update(data map[string]interface{}) UpdateResult {
	ent, err := model.ModelRegister.GetModelByType(m.typEle)
	if err != nil {
		return UpdateResult{
			Error: err,
		}
	}
	fields := []string{}
	setterArsg := []interface{}{}
	for k, v := range data {
		fields = append(fields, fmt.Sprintf("%s=?", k))
		setterArsg = append(setterArsg, v)
	}

	strWhere, argWhere := m.getFilter()
	setterArsg = append(setterArsg, argWhere...)
	sql := "Update " + ent.Entity.TableName + " set " + strings.Join(fields, ",") + " where " + strWhere
	sql, err = internal.OnceCall("modelTypeWhere/"+m.db.DriverName+"/"+sql, func() (string, error) {
		sqlInfo, err := compiler.Compile(sql, m.db.DriverName, false, false)
		if err != nil {
			return "", err
		}
		sqlParser, err := factory.DialectFactory.Create(m.db.DriverName).BuildSql(sqlInfo.Info)
		if err != nil {
			return "", err
		}
		return sqlParser.Sql, nil
	})

	if err != nil {
		return UpdateResult{
			Error: err,
		}
	}
	if m.db.DriverName=="mysql" {
		sql, setterArsg,err=internal.Helper.FixParam(sql, setterArsg)
		if err!=nil {
			return UpdateResult{
				Error: err,
			}
		}
	}
	if Options.ShowSql {
		fmt.Println("-----------------------------")
		fmt.Println(sql)
		fmt.Println("-----------------------------")
	}
	rs, err := m.db.Exec(sql, setterArsg...)
	if err != nil {
		return UpdateResult{
			Error: m.db.parseError(err),
		}
	}
	rn, err := rs.RowsAffected()
	if err != nil {
		return UpdateResult{
			Error: m.db.parseError(err),
		}
	}
	return UpdateResult{
		RowsAffected: rn,
	}

}
