package entity

import (
	"reflect"
	"sync"
)

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
type Entity struct {
	EntityType               reflect.Type
	TableName                string
	Cols                     []ColumnDef          //<-- list of all columns
	MapCols                  map[string]ColumnDef //<-- used for faster access to column by name
	PrimaryConstraints       map[string][]ColumnDef
	UniqueConstraints        map[string][]ColumnDef
	IndexConstraints         map[string][]ColumnDef
	BuildUniqueConstraints   map[string][]ColumnDef
	cacheGetAutoValueColumns sync.Map
	DbTableName              string
}
