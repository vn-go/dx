package dx

import (
	"context"
	"errors"
	"fmt"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
)

func (db *DB) CompilerSqlSelect(sqlSelect string) (*compiler.CompileSelectResult, error) {
	return compiler.CompileSelect(sqlSelect, db.DriverName)
}

// DB is assumed wrapper around *sql.DB with DriverName field and helper methods.
// type DB struct {
//     *sql.DB
//     DriverName string
// }

// ExecRawSqlSelectToDict compiles the abstract SELECT SQL into driver-specific SQL
// using compiler.Compile + DialectFactory, executes it with args, and returns
// all rows as []map[column]value. If no row found, returns an empty slice.
func (db *DB) ExecRawSqlSelectToDict(ctx context.Context, sqlSelect string, args ...any) ([]map[string]any, error) {
	// 1) Compile abstract SQL into dialect-specific SQL
	sqlInfo, err := compiler.Compile(sqlSelect, db.DriverName, false, false)
	if err != nil {
		return nil, fmt.Errorf("compile select: %w", err)
	}
	sqlCompiled, err := factory.DialectFactory.Create(db.DriverName).BuildSql(sqlInfo.Info)
	if err != nil {
		return nil, fmt.Errorf("build sql: %w", err)
	}

	// 2) Ensure context
	if ctx == nil {
		ctx = context.Background()
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
