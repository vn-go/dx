package dx

import (
	"context"
	"database/sql"
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
	Args internal.SqlArgs
}
type datasourceType struct {
	defaultSelector string
	cmpInfo         *compiler.SqlCompilerInfo
	// sqlInfo *types.SqlInfo
	db   *DB
	args internal.SelectorTypesArgs
	// args for function ToSQL
	argsExecutor          internal.SelectorTypesArgs
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
	//strGroupBy      string
	//GroupByExprs []string

	key string
	// autoGroupbyField map[string]string
	aggExpr map[string]exprWithArgs
	// aggExprRevert    map[string]string
	isFormSql     bool
	extraTextArgs []string
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

type datasourceTypeBuildStruct struct {
	DriverName, key, strGroupBy, strWhereOrigin, strSelectOrigin, strWhere string
	isFormSql, skipCheck, whereIsInHaving                                  bool
	OutputFieldsOrExprAlias                                                map[string]types.OutputExpr
	aggExpr                                                                map[string]exprWithArgs
	//strWhereUseAliasField                                                  exprWithArgs
	selector    map[string]bool
	strWhereNew *compiler.CompilerFilterTypeResult
}

func datasourceTypeBuildWhere(ds *datasourceTypeBuildStruct, args *internal.SqlArgs) error {

	if ds.strWhere == "" {
		return nil
	}
	dialect := factory.DialectFactory.Create(ds.DriverName)

	var err error
	if ds.isFormSql {
		ds.strWhereNew, err = compiler.CmpWhere.MakeFilter(
			dialect,
			ds.OutputFieldsOrExprAlias,
			ds.strWhere,
			ds.key, 0,
			0,
			args.Len())
	} else {
		ds.strWhereNew, err = compiler.CmpWhere.MakeFilter(
			dialect,
			ds.OutputFieldsOrExprAlias,
			ds.strWhere, ds.key, 0, 0, args.Len())
	}

	if err != nil {

		return err

	}

	//check if field in fields is agg func expr

	// ds.strWhereUseAliasField.Expr = strWhereNew.FieldExpr
	// ds.strWhereUseAliasField.Args = strWhereNew.Args
	ok := false
	for k := range ds.strWhereNew.Fields {
		if _, ok = ds.aggExpr[k]; ok {
			break
		}
	}
	ds.whereIsInHaving = ok || ds.strWhereNew.HasAggregateFunc
	if ds.whereIsInHaving {
		for k := range ds.strWhereNew.Fields {
			if _, ok = ds.aggExpr[strings.ToLower(k)]; !ok {
				if _, ok := ds.selector[strings.ToLower(k)]; !ok {
					if !ds.skipCheck {
						ds.strGroupBy += "," + k
					} else {
						return compiler.NewCompilerError(fmt.Sprintf("'%s' has field '%s', but not found in '%s", ds.strWhereOrigin, k, ds.strSelectOrigin))

					}

				}

			}
		}

		//ds.args.ArgHaving = append(ds.args.ArgHaving, strWhereNew.Args...)
	}
	return nil
}
func (ds *datasourceType) buildWhere(selectorFieldNotInAggFuns map[string]string, strWhere string, skipCheck bool, startOf2ApostropheArgs, startSqlIndex, startOdDynamicArg int) *compiler.CompilerFilterTypeResult {

	if strWhere == "" {
		return nil
	}
	dialect := factory.DialectFactory.Create(ds.db.DriverName)
	var strWhereNew *compiler.CompilerFilterTypeResult
	var err error
	if ds.isFormSql {
		strWhereNew, err = compiler.CmpWhere.MakeFilter(dialect, ds.cmpInfo.Info.OutputFields, strWhere, ds.key, startOf2ApostropheArgs, startSqlIndex, startOdDynamicArg)
	} else {
		strWhereNew, err = compiler.CmpWhere.MakeFilter(dialect, ds.cmpInfo.Dict.ExprAlias, strWhere, ds.key, startOf2ApostropheArgs, startSqlIndex, startOdDynamicArg)
	}

	if err != nil {
		ds.err = err
		return nil

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
	ds.whereIsInHaving = ok || strWhereNew.HasAggregateFunc
	if ds.whereIsInHaving {
		strGroupByItems := []string{}
		mapGroupByItems := map[string]bool{}
		for k, v := range fields {
			if _, ok = ds.aggExpr[strings.ToLower(k)]; !ok {
				if _, ok := ds.selector[strings.ToLower(k)]; !ok {
					if !skipCheck {
						if _, ok := mapGroupByItems[v.Expr.ExprContent]; !ok {
							strGroupByItems = append(strGroupByItems, v.Expr.ExprContent)
							mapGroupByItems[v.Expr.ExprContent] = true
						}

					} else {
						ds.err = compiler.NewCompilerError(fmt.Sprintf("'%s' has field '%s', but not found in '%s", ds.strWhereOrigin, k, ds.strSelectOrigin))
						return nil
					}

				}

			}
		}
		// for _, x := range selectorFieldNotInAggFuns {
		// 	if _, ok := mapGroupByItems[x]; !ok {
		// 		strWhereNew.GroupByExprs = append(strWhereNew.GroupByExprs, x)
		// 		mapGroupByItems[x] = true
		// 	}
		// }
		strWhereNew.GroupByExprs = strGroupByItems
		// ds.strGroupBy += strings.Join(strGroupByItems, ",")
	}
	return strWhereNew
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
		ds.args.ArgsSelect = append(ds.args.ArgsSelect, args...)
	}

	return ds
}
func (ds *datasourceType) buildSelect(sqlSelect string, strartOf2ApostropheArgs, startOfSqlIndex int) (*compiler.ResolevSelectorResult, error) {
	//

	if sqlSelect == "" {
		return nil, nil
		//selector = ds.defaultSelector
	}
	dialect := factory.DialectFactory.Create(ds.db.DriverName)
	var selectors *compiler.ResolevSelectorResult
	var err error
	if ds.isFormSql {
		selectors, err = compiler.CompilerSelect.MakeSelect(dialect, &ds.cmpInfo.Info.OutputFields, sqlSelect, ds.key, strartOf2ApostropheArgs, startOfSqlIndex)
	} else {
		selectors, err = compiler.CompilerSelect.MakeSelect(dialect, &ds.cmpInfo.Dict.ExprAlias, sqlSelect, ds.key, strartOf2ApostropheArgs, startOfSqlIndex)
	}

	if err != nil {

		return nil, err
	}

	ds.selector = map[string]bool{}
	for _, selector := range selectors.Selectors {
		if selector.FieldExprType != compiler.FieldExprType_Field {
			if selector.Alias == "" {
				//return nil, nil
				ds.err = compiler.NewCompilerError(fmt.Sprintf("'%s' require alias, expression is '%s'", selector.OriginalExpr, sqlSelect))

			}
		}
	}
	if selectors.Selectors.HasAggregateFunction() {
		for _, x := range selectors.Selectors {
			if x.FieldExprType == compiler.FieldExprType_Field {
				// if current selector is agg function call
				selectors.GroupByExprs = append(selectors.GroupByExprs, x.Expr)

				ds.args.ArgGroup = append(ds.args.ArgGroup, x.Args.CompileArgs(ds.args.ArgsSelect, selectors.ApostropheArg)...) // add agrs group by
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
	}

	return selectors, nil

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

func (ds *datasourceTypeSql) fixPostgresParamType() *datasourceTypeSql {

	sql := internal.Helper.FixPostgresParamType(ds.Sql, ds.Args)
	ds.Sql = sql
	return ds
}

type sqlParseStruct struct {
	sqlParse *types.SqlParse
	where    *compiler.CompilerFilterTypeResult
	selector *compiler.ResolevSelectorResult
	Args     internal.SqlArgs
}

func (ds *datasourceType) getSqlParse(startOf2ApostropheArgs, startOfSqlIndex int) (*sqlParseStruct, error) {
	return internal.OnceCall(fmt.Sprintf("datasourceType://ToSql/%s/%s/%s", ds.key, ds.strSelectOrigin, ds.strWhereOrigin), func() (*sqlParseStruct, error) {
		var db = ds.db
		var args internal.SqlArgs = []internal.SqlArg{}
		apostropheArgs := []string{}
		// var ctx = ds.ctx
		var sqlInfo = ds.cmpInfo.Info.Clone()
		defer types.PutSqlInfo(sqlInfo)

		selector, err := ds.buildSelect(ds.strSelectOrigin, startOf2ApostropheArgs, startOfSqlIndex)

		if err != nil {
			return nil, err
		}
		selectorFieldNotInAggFuns := map[string]string{}
		if selector != nil {
			sqlInfo.StrSelect = selector.StrSelectors
			args = append(args, selector.Args...)
			apostropheArgs = append(apostropheArgs, selector.ApostropheArg...)
			for _, x := range selector.Selectors {
				selectorFieldNotInAggFuns = internal.UnionMap(selectorFieldNotInAggFuns, x.FieldMap)
			}

		}

		where := ds.buildWhere(selectorFieldNotInAggFuns, ds.strWhereOrigin, false, len(apostropheArgs), len(args), len(*args.GetDynamicArgs()))

		if ds.err != nil {
			return nil, ds.err
		}
		if where != nil {
			args = append(args, where.Args...)
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
						sqlInfo.StrHaving = ds.strWhere
						//ds.args.ArgHaving = ds.strWhereUseAliasField.Args
					}
					if sqlInfo.StrWhere != "" {
						sqlInfo.StrHaving = "(" + sqlInfo.StrWhere + ") AND (" + sqlInfo.StrHaving + ")"
						ds.args.ArgHaving = append(ds.args.ArgWhere, ds.args.ArgHaving...)
						sqlInfo.StrWhere = ""
						ds.args.ArgWhere = []any{}
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
		groupByExprs := []string{}
		if selector != nil {
			groupByExprs = internal.UnionList(groupByExprs, selector.GroupByExprs)
		}
		if where != nil {
			groupByExprs = internal.UnionList(groupByExprs, where.GroupByExprs)
		}

		if len(groupByExprs) > 0 {
			sqlInfo.StrGroupBy = strings.Join(groupByExprs, ",")
		}
		//internal.UnionList(ds.GroupByExprs, selectors.)
		// if ds.strGroupBy != "" {

		// 	if sqlInfo.StrGroupBy == "" {
		// 		sqlInfo.StrGroupBy = ds.strGroupBy
		// 	} else {
		// 		sqlInfo.StrGroupBy += "," + ds.strGroupBy

		// 	}

		// }

		sqlParse, er := factory.DialectFactory.Create(db.DriverName).BuildSqlNoCache(sqlInfo)
		if er != nil {
			return nil, er
		}
		ret := &sqlParseStruct{
			sqlParse: sqlParse,
			where:    where,
			selector: selector,
			Args:     args,
		}
		return ret, nil

	})
}
func (ds *datasourceType) ToSql() (*datasourceTypeSql, error) {
	apostropheArg := []string{}
	if ds.err != nil {
		return nil, ds.err
	}

	sqlParse, err := ds.getSqlParse(0, 0)
	if err != nil {
		return nil, err
	}
	ds.argsExecutor = ds.args
	if sqlParse.selector != nil {
		ds.strSelect = sqlParse.selector.StrSelectors
		apostropheArg = append(apostropheArg, sqlParse.selector.ApostropheArg...)
		//ds.argsExecutor.ArgsSelect = append(ds.argsExecutor.ArgsSelect, sqlParse.selector.Args...)
	}

	if sqlParse.where != nil {

		if ds.whereIsInHaving {
			if sqlParse.where != nil {
				apostropheArg = append(apostropheArg, sqlParse.where.ApostropheArg...)

			}

		} else {
			if sqlParse.where != nil {
				apostropheArg = append(apostropheArg, sqlParse.where.ApostropheArg...)

			}
		}
	}
	args := []any{}
	if sqlParse.Args.Len() > 0 {
		argsExecutors := ds.argsExecutor.GetArgs(sqlParse.sqlParse.ArgIndex)
		args = sqlParse.Args.CompileArgs(argsExecutors, apostropheArg)
	} else {
		args = ds.argsExecutor.GetArgs(sqlParse.sqlParse.ArgIndex)
	}

	ret := &datasourceTypeSql{
		Sql:  sqlParse.sqlParse.Sql,
		Args: args,
	}
	if ds.db.DriverName == "postgres" {
		ret = ret.fixPostgresParamType()
	}

	return ret, nil

}

func (ds *datasourceType) ToData() (any, error) {
	if ds.err != nil {
		return nil, ds.err
	}

	db := ds.db
	ctx := ds.ctx
	sqlCompiled, err := ds.ToSql()
	if err != nil {
		return nil, compiler.NewCompilerError(err.Error())
	}

	// Đảm bảo có context
	if ctx == nil {
		ctx = context.Background()
	}

	// Fix param cho MySQL (vì placeholder khác PostgreSQL)
	if db.DriverName == "mysql" {
		sqlCompiled.Sql, sqlCompiled.Args, err = internal.Helper.FixParam(sqlCompiled.Sql, sqlCompiled.Args)
		if err != nil {
			return nil, err
		}
	}

	if Options.ShowSql {
		fmt.Println("------------- SQL -------------")
		fmt.Println(sqlCompiled.Sql)
		fmt.Println("------------- ARGS ------------")
		fmt.Printf("%#v\n", sqlCompiled.Args)
		fmt.Println("-------------------------------")
	}

	return ds.db.fecthItemsToSliceVal(sqlCompiled.Sql, ds.ctx, nil, false, sqlCompiled.Args...)
}
func (ds *datasourceType) ToDict() ([]map[string]any, error) {
	if ds.err != nil {
		return nil, ds.err
	}

	db := ds.db
	ctx := ds.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	sqlCompiled, err := ds.ToSql()
	if err != nil {
		return nil, compiler.NewCompilerError(err.Error())
	}

	// Fix MySQL param syntax
	if db.DriverName == "mysql" {
		sqlCompiled.Sql, sqlCompiled.Args, err = internal.Helper.FixParam(sqlCompiled.Sql, sqlCompiled.Args)
		if err != nil {
			return nil, err
		}
	}

	if Options.ShowSql {
		fmt.Println("-------------")
		fmt.Println(sqlCompiled.Sql)
		fmt.Println("-------------")
	}

	// Prepare statement for better performance (DB may reuse plan)
	stmt, err := db.PrepareContext(ctx, sqlCompiled.Sql)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, sqlCompiled.Args...)
	if err != nil {
		errParse := factory.DialectFactory.Create(db.DriverName).ParseError(nil, err)
		if dbErr := Errors.IsDbError(errParse); dbErr != nil && dbErr.ErrorType == Errors.ERR_SYNTAX {
			return nil, compiler.NewCompilerError("Error syntax")
		}
		return nil, fmt.Errorf("query exec: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}
	if len(cols) == 0 {
		return nil, errors.New("no columns returned")
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("colTypes: %w", err)
	}

	// Cache mapping column type → reflect.Type
	cacheKey := ds.db.DriverName + "://" + sqlCompiled.Sql
	colTypeInfo, _ := internal.OnceCall(cacheKey, func() (any, error) {
		return internal.Helper.CreateRowsFromSqlColumnType(sqlCompiled.Sql, colTypes), nil
	})

	typeRowVals := colTypeInfo.([]reflect.Value)

	// Reuse buffers
	dest := make([]any, len(typeRowVals))
	for i := range typeRowVals {
		dest[i] = reflect.New(typeRowVals[i].Type()).Interface()
	}

	results := make([]map[string]any, 0, 64)

	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		rowMap := make(map[string]any, len(cols))
		for i, cn := range cols {
			v := dest[i]
			switch val := v.(type) {
			case *sql.RawBytes:
				rowMap[cn] = string(*val)
			case *[]byte:
				if val == nil {
					rowMap[cn] = nil
				} else {
					rowMap[cn] = string(*val)
				}
			default:
				rowMap[cn] = reflect.ValueOf(v).Elem().Interface()
			}
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return results, nil
}

func (ds *datasourceType) ToDictOld() ([]map[string]any, error) {
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
	if ds.db.DriverName == "mysql" {
		sqlCompiled.Sql, sqlCompiled.Args, err = internal.Helper.FixParam(sqlCompiled.Sql, sqlCompiled.Args)
		if err != nil {
			return nil, err
		}
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
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("columns: %w", err)
	}
	// 5) Iterate over all rows
	var results []map[string]any
	for rows.Next() {

		destVals := internal.Helper.CreateRowsFromSqlColumnType(sqlCompiled.Sql, colTypes) //<-- return []relect.Value
		dest := make([]any, len(colTypes))
		for i, v := range destVals {
			dest[i] = v.Interface()
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		// convert row into map[column]value
		rowMap := make(map[string]any, len(cols))
		for i, cn := range cols {
			//v := dest[i]
			rowMap[cn] = dest[i]

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
		sqlInfo, err = compiler.Compile(sqlSelect, db.DriverName, true, true)
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
		sqlInfo, err = compiler.Compile("select "+strings.Join(strField, ",")+" from "+ent.Entity.TableName, db.DriverName, true, false)
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

	sqlInfo, err := compiler.Compile("select "+defaultInfo.strDefaultSelect+" from "+defaultInfo.ent.TableName, db.DriverName, true, false)

	if err != nil {
		return &datasourceType{
			err: err,
		}
	}

	//argsCollected := sqlInfo.Info.Args.ArgJoin.ToSelectorArgs(args)
	// argsCollected := sqlInfo.Info.Args.ToSelectorArgs(args, sqlInfo.ExtraTextParams)
	key := sqlInfo.Info.GetKey()

	ret := &datasourceType{
		defaultSelector: sqlInfo.Info.StrSelect,
		cmpInfo:         sqlInfo,
		key:             key,
		db:              db,
		// args:            argsCollected,
		//isFormSql:     true,
		extraTextArgs: sqlInfo.ExtraTextParams,
	}
	sqlInfo.Info.FieldArs = *ret.args.GetFields()
	return ret
	// if err != nil {
	// 	return &datasourceType{
	// 		err: err,
	// 	}
	// }
	// key := sqlInfo.Info.GetKey()

	// ret := &datasourceType{
	// 	defaultSelector: strings.Join(defaultInfo.defaultItems, ","),
	// 	cmpInfo:         sqlInfo,
	// 	key:             key,
	// 	db:              db,
	// 	args:            internal.SelectorTypesArgs{},
	// }
	// sqlInfo.Info.FieldArs = *ret.args.GetFields()
	// return ret
}
func (db *DB) DatasourceFromSql(sqlSelect string, args ...any) *datasourceType {

	sqlInfo, err := compiler.Compile(sqlSelect, db.DriverName, true, true)

	if err != nil {
		return &datasourceType{
			err: err,
		}
	}

	//argsCollected := sqlInfo.Info.Args.
	argsCollected := sqlInfo.Info.Args.ToSelectorArgs(args, sqlInfo.ExtraTextParams)
	key := sqlInfo.Info.GetKey()

	ret := &datasourceType{
		defaultSelector: sqlInfo.Info.StrSelect,
		cmpInfo:         sqlInfo,
		key:             key,
		db:              db,
		args:            argsCollected,
		isFormSql:       true,
		extraTextArgs:   sqlInfo.ExtraTextParams,
	}
	sqlInfo.Info.FieldArs = *ret.args.GetFields()
	return ret
}
