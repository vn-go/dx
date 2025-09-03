package postgres

import (
	"fmt"

	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

/*
this function will all Unique index  in Pg database and return a map of table name and column info
return map[<Unique index name>]common.ColumnsInfo, error
struct common.ColumnsInfo  below:

	type common.ColumnsInfo struct {
		TableName string
		Columns   []common.ColumnInfo
	}
	type common.ColumnInfo struct {

			Name string //Db field name

			DbType string //Db field type

			Nullable bool

			Length int
		}
		tenantDB.TenantDB is sql.DB
*/
func (m *MigratorLoaderPostgres) LoadAllUniIndex(db *tenantDB.TenantDB) (map[string]common.ColumnsInfo, error) {
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
			idx.indisunique = true 
			AND idx.indisprimary = false
			AND ns.nspname = 'public'
		ORDER BY 
			i.relname, a.attnum;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	result := make(map[string]common.ColumnsInfo)

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

		column := common.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable,
			Length:   length,
		}

		entry, exists := result[indexName]
		if !exists {
			entry = common.ColumnsInfo{
				TableName: tableName,
				Columns:   []common.ColumnInfo{},
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
