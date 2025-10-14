package mysql

import (
	"fmt"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/migrate/loader/types"
)

func (m *MigratorLoaderMysql) LoadForeignKey(db *db.DB, schema string) ([]types.DbForeignKeyInfo, error) {
	query := `
		SELECT
			LOWER(rc.CONSTRAINT_NAME) CONSTRAINT_NAME ,
			kcu.TABLE_NAME,
			kcu.COLUMN_NAME,
			kcu.REFERENCED_TABLE_NAME,
			kcu.REFERENCED_COLUMN_NAME,
			rc.UPDATE_RULE,
			rc.DELETE_RULE,
			kcu.ORDINAL_POSITION
		FROM
			INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
			JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
				ON rc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
				AND rc.CONSTRAINT_SCHEMA = kcu.CONSTRAINT_SCHEMA
		WHERE
			kcu.TABLE_SCHEMA = DATABASE()
		ORDER BY
			kcu.TABLE_NAME, rc.CONSTRAINT_NAME, kcu.ORDINAL_POSITION
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys: %w", err)
	}
	defer rows.Close()

	type key struct {
		Table string
		Name  string
	}
	temp := make(map[key]*types.DbForeignKeyInfo)

	for rows.Next() {
		var constraintName, tableName, columnName, refTable, refColumn, onUpdate, onDelete string
		var ordinalPos int

		if err := rows.Scan(&constraintName, &tableName, &columnName, &refTable, &refColumn, &onUpdate, &onDelete, &ordinalPos); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key row: %w", err)
		}

		k := key{Table: tableName, Name: constraintName}
		if _, exists := temp[k]; !exists {
			temp[k] = &types.DbForeignKeyInfo{
				ConstraintName: constraintName,
				Table:          tableName,
				RefTable:       refTable,

				Columns:    []string{},
				RefColumns: []string{},
			}
		}

		temp[k].Columns = append(temp[k].Columns, columnName)
		temp[k].RefColumns = append(temp[k].RefColumns, refColumn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	result := make([]types.DbForeignKeyInfo, 0, len(temp))
	for _, v := range temp {
		result = append(result, *v)
	}

	return result, nil
}
