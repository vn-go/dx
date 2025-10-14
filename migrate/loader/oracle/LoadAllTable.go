package oracle

import (
	"fmt"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

// LoadAllTable truy vấn tất cả bảng và cột trong schema hiện tại (Oracle)
// LoadAllTable queries all tables and columns in the current schema (Oracle)
func (m *MigratorOracle) LoadAllTable(db *db.DB, schema string) (map[string]map[string]types.ColumnInfo, error) {
	// Lưu ý: Oracle dùng USER_TAB_COLUMNS thay cho information_schema.columns
	query := `
		SELECT 
			t.table_name,
			t.column_name,
			t.data_type,
			CASE t.nullable WHEN 'Y' THEN 1 ELSE 0 END AS is_nullable,
			COALESCE(t.char_length, t.data_length, 0) AS length
		FROM 
			user_tab_columns t
		ORDER BY 
			t.table_name, t.column_id
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	result := make(map[string]map[string]types.ColumnInfo)

	for rows.Next() {
		var (
			tableName  string
			columnName string
			dataType   string
			isNullable int
			length     int
		)

		if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &length); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		if _, ok := result[tableName]; !ok {
			result[tableName] = make(map[string]types.ColumnInfo)
		}

		result[tableName][columnName] = types.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable == 1,
			Length:   length,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return result, nil
}
