package oracle

import (
	"fmt"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

/*
this function will all primary key in Pg database and return a map of table name and column info
return map[<Primary key constraint name>]types.ColumnsInfo, error
struct ColumnsInfo  below:

	type ColumnsInfo struct {
		TableName string
		Columns   []types.ColumnInfo
	}
	type ColumnInfo struct {

			Name string //Db field name

			DbType string //Db field type

			Nullable bool

			Length int
		}
		tenantDB.TenantDB is sql.DB
*/
func (m *MigratorOracle) LoadAllPrimaryKey(db *db.DB, schema string) (map[string]types.ColumnsInfo, error) {
	// Sử dụng USER_CONSTRAINTS và USER_CONS_COLUMNS (Oracle)
	// USe USER_CONSTRAINTS (Oracle) and USER_CONS_COLUMNS (Oracle)
	query := `
		SELECT 
			uc.constraint_name,
			uc.table_name,
			ucc.column_name,
			c.data_type,
			CASE c.nullable WHEN 'Y' THEN 1 ELSE 0 END AS is_nullable,
			COALESCE(c.char_length, c.data_length, 0) AS length
		FROM 
			user_constraints uc
			JOIN user_cons_columns ucc 
				ON uc.constraint_name = ucc.constraint_name
			JOIN user_tab_columns c
				ON c.table_name = ucc.table_name AND c.column_name = ucc.column_name
		WHERE 
			uc.constraint_type = 'P'
		ORDER BY 
			uc.constraint_name, ucc.position
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	result := make(map[string]types.ColumnsInfo)

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

		column := types.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable,
			Length:   length,
		}

		entry, exists := result[constraintName]
		if !exists {
			entry = types.ColumnsInfo{
				TableName: tableName,
				Columns:   []types.ColumnInfo{},
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
