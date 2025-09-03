package mysql

import (
	"database/sql"
	"fmt"

	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

func (m *MigratorLoaderMysql) LoadAllIndex(db *tenantDB.TenantDB) (map[string]common.ColumnsInfo, error) {
	query := `
		SELECT
			s.INDEX_NAME,
			s.TABLE_NAME,
			s.COLUMN_NAME,
			c.DATA_TYPE,
			c.IS_NULLABLE,
			c.CHARACTER_MAXIMUM_LENGTH
		FROM INFORMATION_SCHEMA.STATISTICS s
		JOIN INFORMATION_SCHEMA.COLUMNS c
			ON s.TABLE_SCHEMA = c.TABLE_SCHEMA
			AND s.TABLE_NAME = c.TABLE_NAME
			AND s.COLUMN_NAME = c.COLUMN_NAME
		WHERE
			s.INDEX_NAME != 'PRIMARY'
			AND s.TABLE_SCHEMA = DATABASE()
		ORDER BY s.INDEX_NAME, s.SEQ_IN_INDEX
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer rows.Close()

	result := make(map[string]common.ColumnsInfo)

	for rows.Next() {
		var indexName, tableName, columnName, dataType, isNullable string
		var charMaxLength sql.NullInt64

		if err := rows.Scan(&indexName, &tableName, &columnName, &dataType, &isNullable, &charMaxLength); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		col := common.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable == "YES",
			Length:   0,
		}
		if charMaxLength.Valid {
			col.Length = int(charMaxLength.Int64)
		}

		if _, exists := result[indexName]; !exists {
			result[indexName] = common.ColumnsInfo{
				TableName: tableName,
				Columns:   []common.ColumnInfo{col},
			}
		} else {
			cols := result[indexName].Columns
			cols = append(cols, col)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}
