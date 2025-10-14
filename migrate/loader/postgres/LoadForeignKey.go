package postgres

import (
	"fmt"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

func (m *MigratorLoaderPostgres) LoadForeignKey(db *db.DB) ([]types.DbForeignKeyInfo, error) {
	query := `
		SELECT
			tc.constraint_name,
			tc.table_name,
			kcu.column_name,
			ccu.table_name AS referenced_table,
			ccu.column_name AS referenced_column
		FROM 
			information_schema.table_constraints AS tc
		JOIN 
			information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.constraint_schema = kcu.constraint_schema
		JOIN 
			information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.constraint_schema = tc.constraint_schema
		WHERE 
			tc.constraint_type = 'FOREIGN KEY'
			AND tc.constraint_schema = 'public'
		ORDER BY 
			tc.constraint_name, kcu.ordinal_position;
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	fkMap := make(map[string]*types.DbForeignKeyInfo)

	for rows.Next() {
		var (
			constraintName string
			tableName      string
			columnName     string
			refTable       string
			refColumn      string
		)

		if err := rows.Scan(&constraintName, &tableName, &columnName, &refTable, &refColumn); err != nil {
			return nil, fmt.Errorf("row scan failed: %w", err)
		}

		if _, exists := fkMap[constraintName]; !exists {
			fkMap[constraintName] = &types.DbForeignKeyInfo{
				ConstraintName: constraintName,
				Table:          tableName,
				RefTable:       refTable,
				Columns:        []string{},
				RefColumns:     []string{},
			}
		}

		info := fkMap[constraintName]
		info.Columns = append(info.Columns, columnName)
		info.RefColumns = append(info.RefColumns, refColumn)
	}

	// Flatten map to slice
	result := make([]types.DbForeignKeyInfo, 0, len(fkMap))
	for _, fk := range fkMap {
		result = append(result, *fk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return result, nil
}
