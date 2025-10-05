package dx

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
)

//	type dataSourceArg struct {
//		ArgWhere   []any
//		ArgsSelect []any
//		ArgJoin    []any
//		ArgGroup   []any
//		ArgHaving  []any
//		ArgOrder   []any
//		ArgSetter  []any
//	}
type exprWithArgs struct {
	Expr string
	Args []any
}
type datasourceType struct {
	defaultSelector string
	cmpInfo         *compiler.SqlCompilerInfo
	// sqlInfo *types.SqlInfo
	db                    *DB
	args                  internal.SelectorTypesArgs
	ctx                   context.Context
	err                   error
	strWhere              string
	strWhereUseAliasField exprWithArgs
	// serve for error message
	strWhereOrigin string
	// tell SQL generate that strWhere must place at "HAVING"
	whereIsInHaving bool
	strSelect       string
	selector        map[string]bool
	// serve for error message
	strSelectOrigin string
	strGroupBy      string

	key string
	// autoGroupbyField map[string]string
	aggExpr map[string]exprWithArgs
	// aggExprRevert    map[string]string
}

func (ds *datasourceType) Sort(strSort string) *datasourceType {
	if ds.err != nil {
		return ds
	}
	ds.cmpInfo.Info.StrOrder = strSort
	return ds
}
func (ds *datasourceType) Limit(limit uint64) *datasourceType {
	if ds.err != nil {
		return ds
	}
	ds.cmpInfo.Info.Limit = &limit
	return ds
}
func (ds *datasourceType) Offset(offset uint64) *datasourceType {
	if ds.err != nil {
		return ds
	}
	ds.cmpInfo.Info.Limit = &offset
	return ds
}
func (ds *datasourceType) Where(strWhere string, args ...any) *datasourceType {
	ds.strWhere = strWhere
	ds.strWhereOrigin = strWhere
	ds.args.ArgWhere = append(ds.args.ArgWhere, args...)
	return ds
}
func (ds *datasourceType) buildWhere(strWhere string) {

	if strWhere == "" {
		return
	}
	dialect := factory.DialectFactory.Create(ds.db.DriverName)

	strWhereNew, err := compiler.CmpWhere.MakeFilter(dialect, ds.cmpInfo.Dict.ExprAlias, strWhere, ds.key)
	if err != nil {
		ds.err = err
		return

	}

	strWhere = strWhereNew.Expr
	fields := strWhereNew.Fields
	//check if field in fields is agg func expr
	ds.strWhere = strWhere
	ds.strWhereUseAliasField.Expr = strWhereNew.FieldExpr
	ds.strWhereUseAliasField.Args = strWhereNew.Args
	ok := false
	for k := range fields {
		if _, ok = ds.aggExpr[k]; ok {
			break
		}
	}
	ds.whereIsInHaving = ok
	if ds.whereIsInHaving {

		for k := range fields {
			if _, ok = ds.aggExpr[strings.ToLower(k)]; !ok {
				if _, ok := ds.selector[strings.ToLower(k)]; !ok {
					ds.err = compiler.NewCompilerError(fmt.Sprintf("'%s' has field '%s', but not found in '%s", ds.strWhereOrigin, k, ds.strSelectOrigin))
					return
				}

			}
		}
		ds.args.ArgHaving = append(ds.args.ArgHaving, strWhereNew.Args...)
	} else {
		ds.args.ArgWhere = append(ds.args.ArgWhere, strWhereNew.Args...)
	}

}

// type datasourceTypeArgs struct {
// 	ArgWhere   []any
// 	ArgsSelect []any
// 	ArgJoin    []any
// 	ArgGroup   []any
// 	ArgHaving  []any
// 	ArgOrder   []any
// 	ArgSetter  []any
// }

func (ds *datasourceType) Select(selector string, args ...any) *datasourceType {

	ds.strSelect = selector
	ds.strSelectOrigin = selector
	if len(args) > 0 {
		ds.args.ArgsSelect = append(ds.args.ArgsSelect, args)
	}

	return ds
}
func (ds *datasourceType) buildSelect(selector string) {

	if selector == "" {
		selector = ds.defaultSelector
	}
	dialect := factory.DialectFactory.Create(ds.db.DriverName)

	selectors, err := compiler.CompilerSelect.MakeSelect(dialect, &ds.cmpInfo.Dict.ExprAlias, selector, ds.key)

	if err != nil {
		ds.err = err
		return
	}
	groupByItems := []string{}
	ds.selector = map[string]bool{}
	for _, x := range selectors.Selectors {
		if !x.IsAggFuncCall {
			// if current selector is agg function call
			groupByItems = append(groupByItems, x.Expr)
			ds.args.ArgGroup = append(ds.args.ArgGroup, x.Args...) // add agrs group by
		} else {
			if ds.aggExpr == nil {
				ds.aggExpr = map[string]exprWithArgs{}
			}
			ds.aggExpr[strings.ToLower(x.Alias)] = exprWithArgs{
				Expr: x.Expr,
				Args: x.Args,
			}

		}
		ds.selector[strings.ToLower(x.Alias)] = true
	}
	ds.strGroupBy = strings.Join(groupByItems, ",")

	ds.strSelect = selectors.StrSelectors
	ds.args.ArgsSelect = append(ds.args.ArgsSelect, selectors.Args.ArgsSelect...)

}
func (ds *datasourceType) WithContext(ctx context.Context) *datasourceType {
	if ds.err != nil {
		return ds
	}
	ds.ctx = ctx
	return ds
}

type datasourceTypeSql struct {
	Sql  string
	Args []any
}

func (ds *datasourceType) ToSql() (*datasourceTypeSql, error) {
	if ds.err != nil {
		return nil, ds.err
	}
	sqlParse, err := internal.OnceCall(fmt.Sprintf("datasourceType://ToSql/%s/%s/%s", ds.key, ds.strSelectOrigin, ds.strWhereOrigin), func() (*types.SqlParse, error) {
		var db = ds.db
		// var ctx = ds.ctx
		var sqlInfo = ds.cmpInfo.Info.Clone()
		defer types.PutSqlInfo(sqlInfo)

		ds.buildSelect(ds.strSelectOrigin)
		if ds.err != nil {
			return nil, ds.err
		}
		ds.buildWhere(ds.strWhereOrigin)
		if ds.err != nil {
			return nil, ds.err
		}
		// var args = ds.args
		if ds.whereIsInHaving {
			if ds.db.DriverName == "mysql" {
				if ds.strWhereUseAliasField.Expr != "" {
					if sqlInfo.StrHaving != "" {
						sqlInfo.StrHaving += " AND (" + ds.strWhereUseAliasField.Expr + ")"
						//ds.args.ArgHaving = append(ds.args.ArgHaving, ds.strWhereUseAliasField.Args...)

					} else {
						sqlInfo.StrHaving = ds.strWhereUseAliasField.Expr
						//ds.args.ArgHaving = ds.strWhereUseAliasField.Args
					}

				}
			} else {
				if ds.strWhere != "" {
					if sqlInfo.StrHaving != "" {
						sqlInfo.StrHaving += " AND (" + ds.strWhere + ")"
					} else {
						sqlInfo.StrHaving = ds.strWhere

					}

				}
			}

		} else {
			if ds.strWhere != "" {
				if sqlInfo.StrWhere != "" {
					sqlInfo.StrWhere += " AND (" + ds.strWhere + ")"
				} else {
					sqlInfo.StrWhere = ds.strWhere
				}

			}
		}

		if ds.strSelect != "" {
			sqlInfo.StrSelect = ds.strSelect
		}
		if ds.strGroupBy != "" {

			if sqlInfo.StrGroupBy == "" {
				sqlInfo.StrGroupBy = ds.strGroupBy
			} else {
				sqlInfo.StrGroupBy += "," + ds.strGroupBy

			}

		}
		return factory.DialectFactory.Create(db.DriverName).BuildSqlNoCache(sqlInfo)
	})
	if err != nil {
		return nil, err
	}
	ret := &datasourceTypeSql{
		Sql:  sqlParse.Sql,
		Args: ds.args.GetArgs(sqlParse.ArgIndex),
	}
	// v := reflect.ValueOf(ds.args)
	//args := ds.args.getArgs(sqlParse.ArgIndex)
	// params := []any{}
	// for _, x := range sqlParse.ArgIndex {
	// 	args := v.FieldByIndex(x.Index).Interface().([]any)
	// 	params = append(params, args)
	// }
	return ret, nil

}
func (ds *datasourceType) ToDict() ([]map[string]any, error) {
	if ds.err != nil {
		return nil, ds.err
	}
	var db = ds.db
	var ctx = ds.ctx
	//var sqlInfo = ds.cmpInfo.Info
	//var args = ds.args

	sqlCompiled, err := ds.ToSql()
	if err != nil {
		return nil, compiler.NewCompilerError(err.Error())
	}

	// 2) Ensure context
	if ctx == nil {
		ctx = context.Background()
	}
	if Options.ShowSql {
		fmt.Println("-------------")
		fmt.Println(sqlCompiled.Sql)
		fmt.Println("-------------")
	}
	// 3) Execute query
	rows, err := db.QueryContext(ctx, sqlCompiled.Sql, sqlCompiled.Args...)
	if err != nil {

		errParse := factory.DialectFactory.Create(db.DriverName).ParseError(nil, err)
		if dbErr := Errors.IsDbError(errParse); dbErr != nil {
			if dbErr.ErrorType == Errors.ERR_SYNTAX {
				return nil, compiler.NewCompilerError("Error syntax")
			}
		}
		return nil, fmt.Errorf("query exec: %w", err)
	}
	defer rows.Close()

	// 4) Get column names
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}
	if len(cols) == 0 {
		return nil, errors.New("no columns returned")
	}

	// 5) Iterate over all rows
	var results []map[string]any
	for rows.Next() {
		// prepare scan destinations
		values := make([]any, len(cols))
		dest := make([]any, len(cols))
		for i := range dest {
			dest[i] = &values[i]
		}

		// scan one row
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		// convert row into map[column]value
		rowMap := make(map[string]any, len(cols))
		for i, cn := range cols {
			v := values[i]
			switch vv := v.(type) {
			case nil:
				rowMap[cn] = nil
			case []byte:
				// normalize []byte → string
				rowMap[cn] = string(vv)
			default:
				rowMap[cn] = vv
			}
		}
		results = append(results, rowMap)
	}

	// 6) Handle iteration error
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return results, nil
}
func (db *DB) NewDataSource(source any, args ...any) *datasourceType {
	var sqlInfo *compiler.SqlCompilerInfo
	var err error

	if sqlSelect, ok := source.(string); ok {
		sqlInfo, err = compiler.Compile(sqlSelect, db.DriverName, true)
	} else {
		typ := reflect.TypeOf(source)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		ent, err := modelRegistry.GetModelByType(typ)
		if err != nil {
			return &datasourceType{
				err: err,
			}
		}
		strField := []string{}
		for _, c := range ent.Entity.Cols {
			strField = append(strField, fmt.Sprintf(c.Name+" "+c.Field.Name))
		}
		sqlInfo, err = compiler.Compile("select "+strings.Join(strField, ",")+" from "+ent.Entity.TableName, db.DriverName, true)
		if err != nil {
			return &datasourceType{
				err: err,
			}
		}
	}

	if err != nil {
		return &datasourceType{
			err: err,
		}
	}
	return &datasourceType{
		cmpInfo: sqlInfo,
		db:      db,
		//args:    args,
	}
}

var pkgPath = reflect.TypeFor[DB]().PkgPath()
var dbTypeFullName = reflect.TypeFor[DB]().Name()

type getDefaultSelectOfModelByModelNameResult struct {
	strDefaultSelect string
	defaultItems     []string
	ent              *entity.Entity
}

func (db *DB) getDefaultSelectOfModelByModelName(modleName string) (*getDefaultSelectOfModelByModelNameResult, error) {
	return internal.OnceCall(fmt.Sprintf("%s/$getDefaultSelectOfModelByModelName", dbTypeFullName), func() (*getDefaultSelectOfModelByModelNameResult, error) {

		// var err error

		ent := modelRegistry.FindEntityByModelName(modleName)
		if ent == nil {
			return nil, compiler.NewCompilerError(fmt.Sprintf("invalid datasource '%s'", modleName))
		}
		strField := []string{}
		defaultItems := []string{}
		for _, c := range ent.Cols {
			strField = append(strField, fmt.Sprintf(c.Name+" "+c.Field.Name))
			defaultItems = append(defaultItems, fmt.Sprintf(c.Field.Name))
		}
		strDefaultSelect := strings.Join(strField, ",")
		return &getDefaultSelectOfModelByModelNameResult{
			strDefaultSelect: strDefaultSelect,
			defaultItems:     defaultItems,
			ent:              ent,
		}, nil
	})
}
func (db *DB) ModelDatasource(modleName string) *datasourceType {

	defaultInfo, err := db.getDefaultSelectOfModelByModelName(modleName)
	if err != nil {
		return &datasourceType{
			err: err,
		}
	}

	sqlInfo, err := compiler.Compile("select "+defaultInfo.strDefaultSelect+" from "+defaultInfo.ent.TableName, db.DriverName, true)

	if err != nil {
		return &datasourceType{
			err: err,
		}
	}
	key := sqlInfo.Info.GetKey()

	ret := &datasourceType{
		defaultSelector: strings.Join(defaultInfo.defaultItems, ","),
		cmpInfo:         sqlInfo,
		key:             key,
		db:              db,
		args:            internal.SelectorTypesArgs{},
	}
	sqlInfo.Info.FieldArs = *ret.args.GetFields()
	return ret
}
