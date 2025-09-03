package postgres

import (
	"fmt"

	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

/*
	 this function will all table in Pg database and return a map of table name and column info

	 #colum info struct

		type common.ColumnInfo struct {

			Name string //Db field name

			DbType string //Db field type

			Nullable bool

			Length int
		}
		tenantDB.TenantDB is sql.DB
*/
func (m *MigratorLoaderPostgres) LoadAllTable(db *tenantDB.TenantDB) (map[string]map[string]common.ColumnInfo, error) {
	query := `
		SELECT 
			table_name, 
			column_name, 
			data_type, 
			is_nullable, 
			COALESCE(character_maximum_length, 0) AS length
		FROM 
			information_schema.columns
		WHERE 
			table_schema = 'public'
		ORDER BY 
			table_name, ordinal_position;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	result := make(map[string]map[string]common.ColumnInfo)

	for rows.Next() {
		var (
			tableName  string
			columnName string
			dataType   string
			isNullable string
			length     int
		)

		if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &length); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		if _, ok := result[tableName]; !ok {
			result[tableName] = make(map[string]common.ColumnInfo)
		}

		result[tableName][columnName] = common.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable == "YES",
			Length:   length,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return result, nil

}
