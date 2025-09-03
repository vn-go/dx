package postgres

import (
	"database/sql"
	"fmt"

	"github.com/vn-go/dx/internal"
)

/*
this function will all  index  in Pg database and return a map of table name and column info
return map[<index name>]internal.ColumnsInfo, error
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

func (m *MigratorLoaderPostgres) LoadAllIndex(db *sql.DB) (map[string]internal.ColumnsInfo, error) {
	query := `
		SELECT
			i.relname AS index_name,
			t.relname AS table_name,
			a.attname AS column_name,
			format_type(a.atttypid, a.atttypmod) AS data_type,
			NOT a.attnotnull AS is_nullable,
			COALESCE(NULLIF(a.atttypmod, -1), 0) AS length
		FROM 
			pg_index idx
		JOIN 
			pg_class i ON i.oid = idx.indexrelid
		JOIN 
			pg_class t ON t.oid = idx.indrelid
		JOIN 
			pg_namespace ns ON ns.oid = t.relnamespace
		JOIN 
			pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(idx.indkey)
		WHERE 
			idx.indisprimary = false
			AND idx.indisunique = false
			AND ns.nspname = 'public'
		ORDER BY 
			i.relname, a.attnum;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	result := make(map[string]internal.ColumnsInfo)

	for rows.Next() {
		var (
			indexName  string
			tableName  string
			columnName string
			dataType   string
			isNullable bool
			length     int
		)

		if err := rows.Scan(&indexName, &tableName, &columnName, &dataType, &isNullable, &length); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		column := internal.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable,
			Length:   length,
		}

		entry, exists := result[indexName]
		if !exists {
			entry = internal.ColumnsInfo{
				TableName: tableName,
				Columns:   []internal.ColumnInfo{},
			}
		}
		entry.Columns = append(entry.Columns, column)
		result[indexName] = entry
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return result, nil
}
