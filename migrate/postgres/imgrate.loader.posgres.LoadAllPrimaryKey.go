package postgres

import (
	"database/sql"
	"fmt"

	"github.com/vn-go/dx/internal"
)

/*
this function will all primary key in Pg database and return a map of table name and column info
return map[<Primary key constraint name>]internal.ColumnsInfo, error
struct internal.ColumnsInfo  below:

	type internal.ColumnsInfo struct {
		TableName string
		Columns   []internal.ColumnInfo
	}
	type internal.ColumnInfo struct {

			Name string //Db field name

			DbType string //Db field type

			Nullable bool

			Length int
		}
		sql.DB is sql.DB
*/
func (m *MigratorLoaderPostgres) LoadAllPrimaryKey(db *sql.DB) (map[string]internal.ColumnsInfo, error) {
	query := `
		SELECT
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			format_type(a.atttypid, a.atttypmod) AS data_type,
			NOT a.attnotnull AS is_nullable,
			COALESCE(NULLIF(a.atttypmod, -1), 0) AS length
		FROM 
			information_schema.table_constraints AS tc
		JOIN 
			information_schema.key_column_usage AS kcu 
			ON tc.constraint_name = kcu.constraint_name 
			AND tc.table_name = kcu.table_name
			AND tc.constraint_schema = kcu.constraint_schema
		JOIN 
			pg_class t ON t.relname = tc.table_name
		JOIN 
			pg_namespace ns ON ns.nspname = tc.constraint_schema
		JOIN 
			pg_attribute a ON a.attrelid = t.oid AND a.attname = kcu.column_name
		WHERE 
			tc.constraint_type = 'PRIMARY KEY' 
			AND tc.constraint_schema = 'public'
		ORDER BY 
			tc.constraint_name, kcu.ordinal_position;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	result := make(map[string]internal.ColumnsInfo)

	for rows.Next() {
		var (
			constraintName string
			tableName      string
			columnName     string
			dataType       string
			isNullable     bool
			length         int
		)

		if err := rows.Scan(&constraintName, &tableName, &columnName, &dataType, &isNullable, &length); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		column := internal.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable,
			Length:   length,
		}

		entry, exists := result[constraintName]
		if !exists {
			entry = internal.ColumnsInfo{
				TableName: tableName,
				Columns:   []internal.ColumnInfo{},
			}
		}
		entry.Columns = append(entry.Columns, column)
		result[constraintName] = entry
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return result, nil
}
