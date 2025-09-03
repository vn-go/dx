package mysql

import (
	"database/sql"
	"fmt"

	"github.com/vn-go/dx/migrate/common"
	"github.com/vn-go/dx/tenantDB"
)

/*
this function will all primary key in MySql database and return a map of table name and column info
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
func (m *MigratorLoaderMysql) LoadAllPrimaryKey(db *tenantDB.TenantDB) (map[string]common.ColumnsInfo, error) {
	query := `
		SELECT
			kcu.CONSTRAINT_NAME,
			kcu.TABLE_NAME,
			kcu.COLUMN_NAME,
			c.DATA_TYPE,
			c.IS_NULLABLE,
			c.CHARACTER_MAXIMUM_LENGTH
		FROM
			INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
				ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
				AND tc.TABLE_SCHEMA = kcu.TABLE_SCHEMA
				AND tc.TABLE_NAME = kcu.TABLE_NAME
			JOIN INFORMATION_SCHEMA.COLUMNS c
				ON c.TABLE_SCHEMA = kcu.TABLE_SCHEMA
				AND c.TABLE_NAME = kcu.TABLE_NAME
				AND c.COLUMN_NAME = kcu.COLUMN_NAME
		WHERE
			tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
			AND tc.TABLE_SCHEMA = DATABASE()
		ORDER BY
			kcu.CONSTRAINT_NAME, kcu.ORDINAL_POSITION
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query primary keys: %w", err)
	}
	defer rows.Close()

	result := make(map[string]common.ColumnsInfo)

	for rows.Next() {
		var constraintName, tableName, columnName, dataType, isNullable string
		var charMaxLength sql.NullInt64

		if err := rows.Scan(&constraintName, &tableName, &columnName, &dataType, &isNullable, &charMaxLength); err != nil {
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
		fakeConstraintName := fmt.Sprintf("%s_%s", constraintName, tableName)
		if _, exists := result[fakeConstraintName]; !exists {
			result[fakeConstraintName] = common.ColumnsInfo{
				TableName: tableName,
				Columns:   []common.ColumnInfo{col},
			}
		} else {
			cols := result[fakeConstraintName].Columns
			cols = append(cols, col)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return result, nil
}
