package dx

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/types"
	dxErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

type selectorTypes struct {
	args      internal.SelectorTypesArgs
	db        *DB
	err       error
	whereExpr *whereTypesItem
	lastWhere *whereTypesItem
	orders    []string
	limit     *uint64
	offset    *uint64
	ctx       context.Context
	sqlTx     *sql.Tx

	selectFields []string
	entityType   *reflect.Type
	valuaOfEnt   reflect.Value
	strJoin      string

	strGroup string

	strHaving string
	strWhere  string
	strSelect string
	strSort   string
}

func (s *selectorTypes) getKey() string {
	if s.selectFields != nil {
		s.strSelect = strings.Join(s.selectFields, ",")
	}

	s.strWhere, s.args.ArgWhere = s.getFilter()
	if s.orders != nil {
		s.strSort = strings.Join(s.orders, ",")
	}

	key := s.strSelect + "+/" + s.strSort + "/" + s.strWhere + "/" + s.strGroup + "/" + s.strHaving + "/" + s.strJoin + "/" + s.strHaving + "/"
	if s.limit != nil {
		key += "/" + fmt.Sprintf("%d", *s.limit)
	}
	if s.offset != nil {
		key += "/" + fmt.Sprintf("%d", *s.offset)
	}

	return key

}

var regexpDBSelectFindPlaceHolder = regexp.MustCompile(`\?`)

func (db *DB) Select(args ...any) *selectorTypes {
	strArgs := []string{}
	for _, a := range args {
		if reflect.TypeOf(a) == reflect.TypeFor[string]() {
			strArgs = append(strArgs, a.(string))
		}
	}
	params := []any{}
	var strFields []string
	if len(args) > 1 {

		// Tìm tất cả các kết quả khớp pattern
		matches := regexpDBSelectFindPlaceHolder.FindAllStringIndex(strings.Join(strArgs, ","), -1)
		params = make([]interface{}, len(matches))
		if len(matches) > 0 {

			offsetVar := len(args) - len(matches)
			for i := range matches {
				params[i] = args[offsetVar+i]
			}
		}
		selectFields := args[0 : len(args)-len(matches)]
		strFields = []string{}
		for _, x := range selectFields {
			if reflect.TypeOf(x) == reflect.TypeFor[string]() {
				strFields = append(strFields, x.(string))
			} else {
				errMsg := "db.Select: invalid selector; field placeholder and argument do not correspond"
				errMsg += "\n"
				for _, x := range args {
					errMsg += fmt.Sprintf("%s", x)
				}
				return &selectorTypes{

					err: dxErrors.NewSysError(errMsg),
				}
			}
		}
	} else {
		strFields = strArgs
	}
	if len(params) == 0 {
		params = nil
	}
	ret := &selectorTypes{
		db:     db,
		orders: []string{},

		selectFields: strFields,
	}
	if len(params) > 0 {
		ret.args = internal.SelectorTypesArgs{
			ArgsSelect: params,
		}
	}
	return ret
}
func (selectors *selectorTypes) Select(args ...any) *selectorTypes {
	strArgs := []string{}
	for _, a := range args {
		if reflect.TypeOf(a) == reflect.TypeFor[string]() {
			strArgs = append(strArgs, a.(string))
		}
	}
	re := regexp.MustCompile(`\?`)

	// Tìm tất cả các kết quả khớp pattern
	matches := re.FindAllStringIndex(strings.Join(strArgs, ","), -1)
	params := make([]interface{}, len(matches))
	if len(matches) > 0 {
		params := make([]interface{}, len(matches))
		offsetVar := len(args) - len(matches)
		for i := range matches {
			params[i] = args[offsetVar+i]
		}
	}
	selectFields := args[0 : len(args)-len(matches)]
	strFields := []string{}
	for _, x := range selectFields {
		if reflect.TypeOf(x) == reflect.TypeFor[string]() {
			strFields = append(strFields, x.(string))
		} else {
			errMsg := "db.Select: invalid selector; field placeholder and argument do not correspond"
			errMsg += "\n"
			for _, x := range args {
				errMsg += fmt.Sprintf("%s", x)
			}
			return &selectorTypes{

				err: dxErrors.NewSysError(errMsg),
			}
		}
	}
	selectors.args.ArgsSelect = append(selectors.args.ArgsSelect, params...)
	selectors.selectFields = append(selectors.selectFields, strFields...)

	return selectors
}
func (selectors *selectorTypes) Where(args ...interface{}) *selectorTypes {
	if selectors.err != nil {
		return selectors
	}
	if len(args) == 0 {
		return &selectorTypes{
			err: fmt.Errorf("(db *DB) Where(<requires at least one argument to be passed>,[list of arguments])"),
		}
	}
	if reflect.TypeOf(args[0]) != reflect.TypeFor[string]() {
		return &selectorTypes{
			err: fmt.Errorf("(db *DB) Where(<argument must be string>),[list of arguments])"),
		}
	}
	if selectors.whereExpr == nil {
		selectors.whereExpr = &whereTypesItem{
			filter: args[0].(string),
			args:   args[1:],
		}
		selectors.lastWhere = selectors.whereExpr
	} else {
		selectors.lastWhere.next = &whereTypesItem{
			filter: args[0].(string),
			args:   args[1:],
		}

	}

	return selectors
}
func (selectors *selectorTypes) Limit(limit uint64) *selectorTypes {
	selectors.limit = &limit
	return selectors
}
func (selectors *selectorTypes) Offset(offset uint64) *selectorTypes {
	selectors.offset = &offset
	return selectors
}
func (selectors *selectorTypes) Order(order string) *selectorTypes {
	selectors.orders = append(selectors.orders, order)
	return selectors
}
func (w *selectorTypes) getFilter() (string, []any) {
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

func (selectors *selectorTypes) GetSQL(typModel reflect.Type) (string, []interface{}, error) {

	key := typModel.String() + "/selectorTypes/GetSQL/" + selectors.getKey()

	selectSql, err := internal.OnceCall(key, func() (*types.SqlParse, error) {
		var err error
		ent, err := model.ModelRegister.GetModelByType(typModel)
		if err != nil {
			return nil, err
		}
		tblExpr := ent.Entity.TableName
		if selectors.strSelect == "" {
			strColms := []string{}
			for _, col := range ent.Entity.Cols {
				strColms = append(strColms, tblExpr+"."+col.Name+" "+col.Field.Name)
			}
			selectors.strSelect = strings.Join(strColms, ",")
		}
		if selectors.strJoin != "" {
			tblExpr = tblExpr + " " + selectors.strJoin
		}
		sqlInfo := &types.SqlInfo{
			Limit:      selectors.limit,
			Offset:     selectors.offset,
			StrSelect:  selectors.strSelect,
			StrWhere:   selectors.strWhere,
			StrHaving:  selectors.strHaving,
			StrOrder:   selectors.strSort,
			StrGroupBy: selectors.strGroup,
			From:       tblExpr,
			FieldArs:   *selectors.args.GetFields(),
		}

		sql, err := compiler.GetSql(sqlInfo, selectors.db.DriverName)
		if err != nil {
			return nil, err
		}

		return sql, nil

	})
	if err != nil {
		return "", nil, err
	}
	// internal.UnionMap(selectSql.Args, selectors.args)

	retArgs := selectors.args.GetArgs(selectSql.ArgIndex)
	selectorArg := selectSql.Args.ToSelectorArgs(retArgs, selectSql.ApstropheArgs)
	retArgs = selectorArg.GetArgs(selectSql.ArgIndex)
	return selectSql.Sql, retArgs, err
}

func (selectors *selectorTypes) Find(item any) error {

	if selectors.strJoin != "" {
		return selectors.findByJoin(item)
	} else {
		typeEle := reflect.TypeOf(item)
		if typeEle.Kind() == reflect.Ptr {
			typeEle = typeEle.Elem()
		}
		if typeEle.Kind() != reflect.Slice {
			return dxErrors.NewSysError(fmt.Sprintf("%s is not slice", reflect.TypeOf(item).String()))
		}
		typeEle = typeEle.Elem()
		if selectors.entityType != nil {
			sqlQuery, args, err := selectors.GetSQL(*selectors.entityType)
			if selectors.db.DriverName == "mysql" {
				sqlQuery, args, err = internal.Helper.FixParam(sqlQuery, args)
				if err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
			return selectors.db.fecthItems(item, sqlQuery, selectors.ctx, selectors.sqlTx, true, args...)
		} else {
			//"reflect.Value.Elem"
			sqlQuery, args, err := selectors.GetSQL(typeEle)
			if err != nil {
				return err
			}
			if selectors.db.DriverName == "mysql" {
				sqlQuery, args, err = internal.Helper.FixParam(sqlQuery, args)
				if err != nil {
					return err
				}
			}
			return selectors.db.fecthItems(item, sqlQuery, selectors.ctx, selectors.sqlTx, true, args...)
		}
	}

}
func (selectors *selectorTypes) GetModelType() *reflect.Type {
	return selectors.entityType
}
func (selectors *selectorTypes) First(item any) error {
	typeEle := reflect.TypeOf(item)
	if typeEle.Kind() == reflect.Ptr {
		typeEle = typeEle.Elem()
	}

	selectors.Limit(1)
	if selectors.entityType != nil {
		sqlQuery, args, err := selectors.GetSQL(*selectors.entityType)
		if err != nil {
			return err
		}
		return selectors.db.fecthItem(item, sqlQuery, selectors.ctx, selectors.sqlTx, true, args...)
	} else {
		sqlQuery, args, err := selectors.GetSQL(typeEle)
		if err != nil {
			return err
		}
		return selectors.db.fecthItem(item, sqlQuery, selectors.ctx, selectors.sqlTx, true, args...)
	}

}
