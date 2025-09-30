package dx

import (
	"context"
	"errors"
	"fmt"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
)

type datasourceType struct {
	cmpInfo *compiler.SqlCompilerInfo
	// sqlInfo *types.SqlInfo
	db   *DB
	args []any
	ctx  context.Context
	err  error
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
	var numOfWhereParams = 0
	strWhereNew, err := compiler.CmpWhere.MakeFilter(dialect, ds.cmpInfo.Dict.ExprAlias, strWhere, &numOfWhereParams)
	if err != nil {
		ds.err = err
		return ds
	}

	if ds.cmpInfo.Info.StrWhere != "" {
		ds.cmpInfo.Info.StrWhere += " AND (" + strWhereNew + ")"
	} else {
		ds.cmpInfo.Info.StrWhere = strWhereNew

	}
	if ds.args == nil {
		ds.args = []any{}
	}
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
func (ds *datasourceType) ToDict() ([]map[string]any, error) {
	if ds.err != nil {
		return nil, ds.err
	}
	var db = ds.db
	var ctx = ds.ctx
	var sqlInfo = ds.cmpInfo.Info
	var args = ds.args

	sqlCompiled, err := factory.DialectFactory.Create(db.DriverName).BuildSql(sqlInfo)
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
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
func (db *DB) NewDataSource(sqlSelect string, args ...any) *datasourceType {

	sqlInfo, err := compiler.Compile(sqlSelect, db.DriverName, true)
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
