package mysql

import (
	"database/sql"
	"fmt"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

// LoadAllUniIndex retrieves all unique indexes (excluding primary keys) from the MySQL database
// and organizes them into a map where the key is the index name and the value contains
// the table name and a list of column details for that index.
// Parameters:
//   - db: A pointer to a TenantDB instance for executing database queries.
//
// Returns:
//   - A map[string]types.ColumnsInfo containing unique index details, where the key is the index name.
//   - An error if the query or row scanning fails.
func (m *MigratorLoaderMysql) LoadAllUniIndex(db *db.DB) (map[string]types.ColumnsInfo, error) {
	// SQL query to fetch unique index information from INFORMATION_SCHEMA.
	// Joins STATISTICS and COLUMNS tables to get index name, table name, column name,
	// data type, nullability, and character length for unique indexes (NON_UNIQUE = 0)
	// excluding primary keys, for the current database schema.
	query := `
		SELECT
			LOWER(s.INDEX_NAME) INDEX_NAME,
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
			s.NON_UNIQUE = 0
			AND s.INDEX_NAME != 'PRIMARY'
			AND s.TABLE_SCHEMA = DATABASE()
		ORDER BY s.INDEX_NAME, s.SEQ_IN_INDEX
	`

	// Execute the query on the provided database connection.
	rows, err := db.Query(query)
	if err != nil {
		// Return an error if the query execution fails, wrapping it with context.
		return nil, fmt.Errorf("failed to query unique indexes: %w", err)
	}
	// Ensure rows are closed after use to prevent resource leaks.
	defer rows.Close()

	// Initialize a map to store the results, with index names as keys and ColumnsInfo as values.
	result := make(map[string]types.ColumnsInfo)

	// Iterate over each row returned by the query.
	for rows.Next() {
		// Variables to store scanned row data.
		var indexName, tableName, columnName, dataType, isNullable string
		var charMaxLength sql.NullInt64

		// Scan the row into variables; return an error if scanning fails.
		if err := rows.Scan(&indexName, &tableName, &columnName, &dataType, &isNullable, &charMaxLength); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a ColumnInfo struct with column details.
		col := types.ColumnInfo{
			Name:     columnName,
			DbType:   dataType,
			Nullable: isNullable == "YES",
			Length:   0,
		}
		// Set the column length if the character maximum length is valid.
		if charMaxLength.Valid {
			col.Length = int(charMaxLength.Int64)
		}

		// Check if the index already exists in the result map.
		if _, exists := result[indexName]; !exists {
			// If it doesn't exist, initialize a new ColumnsInfo entry with the table name and column.
			result[indexName] = types.ColumnsInfo{
				TableName: tableName,
				Columns:   []types.ColumnInfo{col},
			}
		} else {
			// If it exists, append the new column to the existing ColumnsInfo's Columns slice.
			cols := result[indexName].Columns
			cols = append(cols, col)

		}
	}

	// Check for any errors that occurred during row iteration.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Return the populated map and nil error on success.
	return result, nil
}
