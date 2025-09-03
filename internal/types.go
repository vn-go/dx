package internal

import (
	"database/sql"
	"reflect"
)

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

type ColumnDef struct {
	/*
		Name of the column in the database.
		If `db:"column:user_name"` is provided, this value will be "user_name".
		Otherwise, default to SnakeCase of field name.
	*/
	Name string

	/*
		SQL data type in target RDBMS. Derived from `type:"..."` tag or Go type mapping.
		Example: "varchar(255)", "int", "jsonb", etc.
	*/
	Type string

	/*
		Whether the column allows NULL values.
		Automatically derived from Go pointer type: e.g., *string → Nullable=true.
	*/
	Nullable bool

	/*
		If tag looks like:
			`db:"pk"` or `db:"primary"` → PKName = Name
			`db:"pk(my_pk_name)"` → PKName = "my_pk_name"
	*/
	PKName string

	/*
		Whether this column is auto-increment.
		Derived from tag `db:"auto"`
	*/
	IsAuto bool

	/*
		Default value for the column. Derived from tag `db:"default(...)"`
		Example: `default(now)` → Default = "now"
	*/
	Default string

	/*
		If tag looks like:
			`db:"unique"` → UniqueName = Name
			`db:"unique(email_uk)"` → UniqueName = "email_uk"
	*/
	UniqueName string

	/*
		If tag looks like:
			`db:"index"` → IndexName = Name
			`db:"index(email_idx)"` → IndexName = "email_idx"
	*/
	IndexName string

	/*
		Length of the column. Used in varchar(n), nvarchar(n), etc.
		Example: type:"string(100)" → Length = 100
	*/
	Length *int

	/*
		Precision and Scale. Used for types like decimal(p,s).
		Example: type:"decimal(10,2)" → Precision=10, Scale=2
	*/
	Precision *int
	Scale     *int

	/*
		Whether this column should be treated as a JSON/document field.
		Derived from type like "json", "jsonb", or Go struct tag/type analysis.
	*/
	IsJSON bool

	/*
		Optional comment or description of the column.
		Derived from: db:"comment(This is the user email)"
	*/
	Comment string

	/*
		Original Go field reference for further metadata.
	*/
	Field reflect.StructField

	/*
		Optional: the referenced table and column if this is a foreign key.
		Derived from: db:"foreign(users.id)" → RefTable = "users", RefColumn = "id"
	*/
	IsForeignKey bool
	RefTable     string
	RefColumn    string
	IndexOfField []int
}

// Certain functions are transformed during compilation based on the SQL dialect.
// The dialect decides whether a function needs to be adapted for the target database driver.
// If required, the function will be rewritten here.
type DialectDelegateFunction struct {
	FuncName         string
	Args             []string
	HandledByDialect bool // ✅ Indicates if this function is allowed to be delegated to the dialect
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
type Dialect interface {
	LikeValue(val string) string

	ParseError(dbSchame *DbSchema, err error) error
	Name() string
	Quote(str ...string) string
	GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error)
	ToText(value string) string
	ToParam(index int) string
	SqlFunction(delegator *DialectDelegateFunction) (string, error)
	MakeSqlInsert(tableName string, columns []ColumnDef, data interface{}) (string, []any)
	NewDataBase(db *sql.DB, sampleDsn string, dbName string) (string, error)
	MakeSelectTop(sql string, top int) string
}
