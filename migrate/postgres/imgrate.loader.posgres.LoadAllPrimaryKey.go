package postgres

import (
	"fmt"

	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

/*
this function will all primary key in Pg database and return a map of table name and column info
return map[<Primary key constraint name>]common.ColumnsInfo, error
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
func (m *MigratorLoaderPostgres) LoadAllPrimaryKey(db *tenantDB.TenantDB) (map[string]common.ColumnsInfo, error) {
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

	result := make(map[string]common.ColumnsInfo)

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

		column := common.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable,
			Length:   length,
		}

		entry, exists := result[constraintName]
		if !exists {
			entry = common.ColumnsInfo{
				TableName: tableName,
				Columns:   []common.ColumnInfo{},
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
