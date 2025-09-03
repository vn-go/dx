package common

import (
	"database/sql"

	"github.com/vn-go/dx/internal"
)

type IMigratorLoader interface {
	GetDbName(db *sql.DB) string
	LoadAllTable(db *sql.DB) (map[string]map[string]internal.ColumnInfo, error)
	LoadAllPrimaryKey(db *sql.DB) (map[string]internal.ColumnsInfo, error)
	/*
		Heed: for SQL Server, we need to use the following query to get the unique keys:
			SELECT
			t.name AS TableName,
			i.name AS IndexName
			FROM sys.indexes i
			JOIN sys.tables t ON i.object_id = t.object_id
			WHERE i.type_desc = 'NONCLUSTERED' and is_unique_constraint=1
	*/
	LoadAllUniIndex(db *sql.DB) (map[string]internal.ColumnsInfo, error)
	/*

	 */
	LoadAllIndex(db *sql.DB) (map[string]internal.ColumnsInfo, error)
	LoadFullSchema(db *sql.DB) (*internal.DbSchema, error)
	LoadForeignKey(db *sql.DB) ([]internal.DbForeignKeyInfo, error)
}
