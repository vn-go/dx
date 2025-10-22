package sql

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

/*
Args of sql query after compiled
*/
type argument struct {
}
type arguments []argument
type contentResullt struct {
}
type refFieldInfo struct {
	RefTable string
	RefField string
}
type refFields []refFieldInfo
type compilerResult struct {
	Content string
	Args    []any
	Fields  refFields
}
type dictionaryField struct {
	Expr string
	Typ  sqlparser.ValType
}
type dictionary struct {
	fields     map[string]dictionaryField
	tableAlias map[string]string
}

type injector struct {
	dict        *dictionary
	dialect     types.Dialect
	fields      refFields
	textParams  []string
	dynamicArgs []any
}

/*
Create a new dictionary
*/
func newDictionary() *dictionary {
	return &dictionary{
		fields:     make(map[string]dictionaryField),
		tableAlias: make(map[string]string),
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
