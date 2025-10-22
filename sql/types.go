package sql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/sqlparser"
)

/*
Args of sql query after compiled
*/
type argument struct {
}
type arguments []argument

/*
use check permission
if user has permission to access this table and field
*/
type refFieldInfo struct {
	EntityName      string
	EntityFieldName string
}

/*
use check permission
if user has permission to access this table and field
*/
type refFields map[string]refFieldInfo
type compilerResult struct {
	Content string
	Args    []any
	/*
		use check permission
		if user has permission to access this table and field

		key: dataset name+"."+ colunm name

		value: refFieldInfo
		Note: dataset name is the name of entity struct, colunm name is the name of field in entity struct
	*/
	Fields refFields
	/*
		after compiled to sql, we need to know the type of each selector in the result set.

		key is combination of dataset name and column name in lower case expression after compiled.

		value is dictionaryField

	*/
	selectedExprs dictionaryFields

	/*
		Example:

			SELECT ID, Name, COUNT(ID) as count FROM T1 Where count>10 and type='aa'

			Determine count in where clause by using selectedExprsReverse

			selectedExprsReverse["t1.count"] = "COUNT(ID)"

			Complied Sql is:

				SELECT ID, Name, COUNT(ID) as count FROM T1 having "COUNT(ID)" group by name Where  type='aa'
	*/
	selectedExprsReverse dictionaryFields // reverse of selectedExprs

	/*
		after compiled to sql, all fields in select, where anf fform clase will be stored.

		key is combination of dataset name and column name in lower case <dataset name>.<column name>

		value is dictionaryField

		Note: store field only. When build group by and having clause,
		we need to know the type of each field in the group by clause.
		Field in group by clause is a field not in aggregate function.
		if where clause has aggregate function it will be placed in having clause.
		when compile where clause. where clause is list of expression separated by AND (only).




		Example: "COUNT(ID) > 10 and code='abc'" -> COUNT(ID) > 10,code='abc'

		--> having clause: COUNT(ID) > 10
		--> where clause: code='abc'


	*/
	//allFields dictionaryFields
}

// After compiled to sql, we need to know the type of each field in the result set.
type dictionaryField struct {
	/*
		Expression of select field
		Example: "[T1].[ID]" -> assume that dialect is MSSQL
	*/
	Expr              string
	Typ               sqlparser.ValType
	IsInAggregateFunc bool
	Alias             string
	EntityField       *entity.ColumnDef
}

// this type is very important for make dynamic struct from column types
// for db.rows.Scan()
// ref file: sql/dictionaryFields.toDynamicStruct.go
type dictionaryFields map[string]dictionaryField

type dictionary struct {
	fields dictionaryFields
	/*
	 map dataset name and alias (alias in in from clause)
	 and also map database table name and alias (alias in in from clause)
	*/
	tableAlias map[string]string
	/*
	 map

	 key: strings.ToLower(ent.EntityType.Name())

	 value: *ent
	*/
	entities map[string]*entity.Entity

	/*
		map alias to entity

		key: alias or table name in from clause

		value: *ent


	*/
	aliasToEntity map[string]*entity.Entity
}

type injector struct {
	dict    *dictionary
	dialect types.Dialect
	/*
		use check permission
		if user has permission to access this table and field
	*/
	fields      refFields
	textParams  []string
	dynamicArgs []any
}

/*
Create a new dictionary
*/
func newDictionary() *dictionary {
	return &dictionary{
		fields:        make(map[string]dictionaryField),
		tableAlias:    make(map[string]string),
		entities:      make(map[string]*entity.Entity),
		aliasToEntity: make(map[string]*entity.Entity),
	}
}
func newInjector(dialect types.Dialect, textParam []string, dynamicArgs []any) *injector {
	return &injector{
		dict:        newDictionary(),
		dialect:     dialect,
		fields:      refFields{},
		textParams:  textParam,
		dynamicArgs: dynamicArgs,
	}

}

const GET_PARAMS_FUNC = "dx__GetParams"

type CompilerError struct {
	Message string
	Args    []any
}

func (e *CompilerError) Error() string {
	return fmt.Sprintf(e.Message, e.Args...)
}
func newCompilerError(message string, args ...any) *CompilerError {
	return &CompilerError{
		Message: message,
		Args:    args,
	}
}

type sqlComplied struct {
	source   string // from
	selector string // select field here
	filter   string // where clause
	sort     string // order by clause
}

func (s *sqlComplied) String() string {
	query := "SELECT "

	// SELECT fields
	if s.selector != "" {
		query += s.selector
	} else {
		query += "*"
	}

	// FROM clause
	if s.source != "" {
		query += " FROM " + s.source
	}

	// WHERE clause
	if s.filter != "" {
		query += " WHERE " + s.filter
	}

	// ORDER BY clause
	if s.sort != "" {
		query += " ORDER BY " + s.sort
	}

	return query
}
