package types

import "github.com/vn-go/dx/db"

type ColumnInfo struct {
	/*
		Db field name
	*/
	Name string
	/*
		Go field name
	*/
	DbType string
	/*
		Is allow null?
	*/
	Nullable bool
	/*
		Length is the length of the column
	*/
	Length int
}
type ColumnsInfo struct {
	TableName string
	Columns   []ColumnInfo
}

/*
This struct is used to store the foreign key information from the database .
*/
type DbForeignKeyInfo struct {
	/**/
	ConstraintName string
	Table          string
	Columns        []string
	RefTable       string
	RefColumns     []string
}
type DbSchema struct {
	/*
		Database name
	*/
	DbName string
	/*
		map[<table name>]map[<column name>]bool
	*/
	Tables map[string]map[string]bool
	/*
		map[<primary key constraint name>]ColumnsInfo
	*/
	PrimaryKeys map[string]ColumnsInfo
	/*
		map[<Unique Keys constraint name>]ColumnsInfo
	*/
	UniqueKeys map[string]ColumnsInfo
	/*
		map[<Index name>]ColumnsInfo
	*/
	Indexes     map[string]ColumnsInfo
	ForeignKeys map[string]DbForeignKeyInfo
}
type IMigratorLoader interface {
	LoadAllTable(db *db.DB) (map[string]map[string]ColumnInfo, error)
	LoadAllPrimaryKey(db *db.DB) (map[string]ColumnsInfo, error)
	/*
		Heed: for SQL Server, we need to use the following query to get the unique keys:
			SELECT
			t.name AS TableName,
			i.name AS IndexName
			FROM sys.indexes i
			JOIN sys.tables t ON i.object_id = t.object_id
			WHERE i.type_desc = 'NONCLUSTERED' and is_unique_constraint=1
	*/
	LoadAllUniIndex(db *db.DB) (map[string]ColumnsInfo, error)
	/*

	 */
	LoadAllIndex(db *db.DB) (map[string]ColumnsInfo, error)
	LoadFullSchema(db *db.DB) (*DbSchema, error)
	LoadForeignKey(db *db.DB) ([]DbForeignKeyInfo, error)
}

var SkipLoadSchemaOnMigrate bool
