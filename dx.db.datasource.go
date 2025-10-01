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
)

type dataSourceArg struct {
	ArgWhere   []any
	ArgsSelect []any
	ArgJoin    []any
	ArgGroup   []any
	ArgHaving  []any
	ArgOrder   []any
	ArgSetter  []any
}
type datasourceType struct {
	cmpInfo *compiler.SqlCompilerInfo
	// sqlInfo *types.SqlInfo
	db        *DB
	args      []any
	ctx       context.Context
	err       error
	strWhere  string
	strSelect string
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
	if ds.err != nil {
		return ds
	}

	dialect := factory.DialectFactory.Create(ds.db.DriverName)

	strWhereNew, err := compiler.CmpWhere.MakeFilter(dialect, ds.cmpInfo.Dict.ExprAlias, strWhere, ds.cmpInfo.Info.GetKey())
	if err != nil {
		ds.err = err
		return ds
	}
	ds.strWhere = strWhereNew
	// if ds.cmpInfo.Info.StrWhere != "" {
	// 	ds.cmpInfo.Info.StrWhere += " AND (" + strWhereNew + ")"
	// } else {
	// 	ds.cmpInfo.Info.StrWhere = strWhereNew

	// }
	if ds.args == nil {
		ds.args = []any{}
	}
	ds.args = append(ds.args, args...)

	return ds
}

func (ds *datasourceType) Select(selector string, args ...any) *datasourceType {
	if ds.err != nil {
		return ds
	}
	dialect := factory.DialectFactory.Create(ds.db.DriverName)

	strSelect, err := compiler.CompilerSelect.MakeSelect(dialect, &ds.cmpInfo.Dict.ExprAlias, selector, ds.cmpInfo.Info.GetKey())

	if err != nil {
		ds.err = err
		return ds
	}
	//ds.cmpInfo.Info.StrSelect = strSelect
	ds.strSelect = strSelect
	ds.args = append(ds.args, args...)
	return ds
}
func (ds *datasourceType) WithContext(ctx context.Context) *datasourceType {
	if ds.err != nil {
		return ds
	}
	ds.ctx = ctx
	return ds
}

func (ds *datasourceType) ToSql() (*types.SqlParse, error) {
	if ds.err != nil {
		return nil, ds.err
	}
	var db = ds.db
	// var ctx = ds.ctx
	var sqlInfo = ds.cmpInfo.Info
	oldStrWhere := ds.cmpInfo.Info.StrWhere
	oldStrSelect := ds.cmpInfo.Info.StrSelect
	defer func() {
		ds.cmpInfo.Info.StrWhere = oldStrWhere
		ds.cmpInfo.Info.StrSelect = oldStrSelect
	}()
	// var args = ds.args
	if ds.strWhere != "" {
		ds.cmpInfo.Info.StrWhere += " AND (" + ds.strWhere + ")"
	}
	if ds.strSelect != "" {
		ds.cmpInfo.Info.StrSelect = ds.strSelect
	}

	return factory.DialectFactory.Create(db.DriverName).BuildSql(sqlInfo)
	// if err != nil {
	// 	return nil, compiler.NewCompilerError(err.Error())
	// }
}
func (ds *datasourceType) ToDict() ([]map[string]any, error) {
	if ds.err != nil {
		return nil, ds.err
	}
	var db = ds.db
	var ctx = ds.ctx
	//var sqlInfo = ds.cmpInfo.Info
	var args = ds.args

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
	rows, err := db.QueryContext(ctx, sqlCompiled.Sql, args...)
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
		args:    args,
	}
}
func (db *DB) ModelDatasource(modleName string) *datasourceType {

	var err error

	ent := modelRegistry.FindEntityByName(modleName)
	if ent == nil {
		return &datasourceType{
			err: compiler.NewCompilerError(fmt.Sprintf("invalid datasource '%s'", modleName)),
		}
	}
	strField := []string{}
	for _, c := range ent.Cols {
		strField = append(strField, fmt.Sprintf(c.Name+" "+c.Field.Name))
	}
	sqlInfo, err := compiler.Compile("select "+strings.Join(strField, ",")+" from "+ent.TableName, db.DriverName, true)
	if err != nil {
		return &datasourceType{
			err: err,
		}
	}
	return &datasourceType{
		cmpInfo: sqlInfo,
		db:      db,
		args:    []any{},
	}
}
