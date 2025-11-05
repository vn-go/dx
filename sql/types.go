package sql

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

/*
Args of sql query after compiled
*/
type argument struct {
	/*
		if const value passed to sql query, it will be stored here.
	*/
	val any
	/*
	 if dynamic arg this store the index of dynamic arg in args array.
	  if not dynamic arg this store 0


	*/
	index int
}
type arguments []argument

func (a arguments) ToArray(dynamicArgs []any) ([]any, error) {
	ret := make([]any, len(a))
	for i, arg := range a {
		if arg.index > 0 {
			if arg.index > len(dynamicArgs) {
				return nil, fmt.Errorf("dynamic arg index out of range. index: %d, dynamic args length: %d", arg.index, len(dynamicArgs))
			}
			ret[i] = dynamicArgs[arg.index-1]
		} else {
			ret[i] = arg.val
		}
	}
	return ret, nil
}

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

func (r refFields) String() any {
	bff, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf("refFields error:%s", err.Error())
	}
	return string(bff)
}

func (r *refFields) merge(fields refFields) refFields {
	*r = internal.UnionMap(*r, fields)
	return *r
}

type outputField struct {
	Name         string
	IsCalculated bool
	FieldType    reflect.Type
	DbType       sqlparser.ValType
}
type outputFields []outputField

func (o *outputFields) ToHas256Key() string {
	if o == nil || len(*o) == 0 {
		return ""
	}

	// Dùng strings.Builder thay vì []string + strings.Join để tránh alloc trung gian
	var b strings.Builder
	for i, f := range *o {
		if i > 0 {
			b.WriteByte(',') // phân cách
		}
		b.WriteString(f.Name)
		b.WriteByte('/')
		b.WriteString(f.FieldType.String())
	}

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

func (o *outputFields) toCamelCase(name string) string {
	if name == "" {
		return name
	}

	runes := []rune(name)

	// Nếu 2 ký tự đầu viết hoa (vd: APIResponse), thì hạ tất cả cho đến khi gặp chữ thường
	// Ví dụ: "APIResponse" -> "apiResponse", "HTTPStatus" -> "httpStatus"
	for i := 0; i < len(runes); i++ {
		if i == 0 {
			runes[i] = unicode.ToLower(runes[i])
		} else if i+1 < len(runes) && unicode.IsUpper(runes[i]) && unicode.IsUpper(runes[i+1]) {
			runes[i] = unicode.ToLower(runes[i])
		} else {
			runes[i] = unicode.ToLower(runes[i])
			break
		}
	}
	return string(runes)
}

type initToStruct struct {
	val  reflect.Type
	once sync.Once
}

var initToStructCache sync.Map

func (o *outputFields) ToStruct(key string) reflect.Type {
	a, _ := initToStructCache.LoadOrStore(key, &initToStruct{})
	i := a.(*initToStruct)
	i.once.Do(func() {
		i.val = o.ToStructNoCache()
	})
	return i.val
}
func (o *outputFields) ToStructNoCache() reflect.Type {
	var fields []reflect.StructField

	for _, f := range *o {
		fieldType := f.FieldType
		if fieldType == nil {
			fieldType = reflect.TypeOf((*interface{})(nil)).Elem() // interface{}
		}

		fields = append(fields, reflect.StructField{
			Name: f.Name,
			Type: fieldType,
			Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, o.toCamelCase(f.Name))),
		})
	}

	return reflect.StructOf(fields)
}
func (o *outputFields) ToArrayOfStruct(key string) reflect.Type {
	return reflect.SliceOf(o.ToStruct(key))
}
func (o outputFields) String() string {
	bff, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Sprintf("outputFields error:%s", err.Error())
	}
	return string(bff)
}

type accessScopes map[string]map[string]string
type compilerResult struct {
	// limit                       string
	// offset                      string
	// index of limit and offset in original args array, before compiled
	reIndex []int
	// orderBy                     string
	IsInSubquery bool
	// use for error message. Error message should be show with original content
	OriginalContent string
	Content         string
	AliasOfContent  string
	Args            arguments
	IsExpression    bool // true if not select field
	/*
		use check permission
		if user has permission to access this table and field

		key: dataset name+"."+ colunm name

		value: refFieldInfo
		Note: dataset name is the name of entity struct, colunm name is the name of field in entity struct
	*/
	Fields             refFields
	AccessScope        accessScopes
	Hash256AccessScope string
	/*
		after compiled to sql, we need to know the type of each selector in the result set.

		key is combination of dataset name and column name in lower case expression after compiled.

		value is dictionaryField

	*/
	selectedExprs dictionaryFields
	// all fields in select expression
	// just field no expresion or function call
	// use to detect group by clause if any aggregate function in select
	nonAggregateFields dictionaryFields
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
	IsInAggregateFunc   bool
	OutputFields        outputFields
	Hash256OutputFields string
	ResultType          reflect.Type
	ResultDbType        sqlparser.ValType
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
	//Children          *dictionaryFields
	Node sqlparser.SQLNode
}

// this type is very important for make dynamic struct from column types
// for db.rows.Scan()
// ref file: sql/dictionaryFields.toDynamicStruct.go
type dictionaryFields map[string]*dictionaryField

// func (d dictionaryFields) String() string {
// 	items := []string{
// 		"DictionaryFields",
// 	}
// 	for k, v := range d {
// 		items = append(items, fmt.Sprintf("%s\t\t\t%s\t\t\t%s\t\t\t%t", k, v.Expr, v.Alias, v.IsInAggregateFunc))
// 	}
// 	return strings.Join(items, "\n")
// }

func (d *dictionaryFields) merge(exprs dictionaryFields) *dictionaryFields {
	*d = internal.UnionMap(*d, exprs)
	return d
}

type subqueryEntityField struct {
	source string
	field  string
}
type subqueryEntityFields map[string]subqueryEntityField
type subqueryEntity struct {
	fields subqueryEntityFields
}
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
	aliasToEntity   map[string]*entity.Entity
	subqueryEntites map[string]subqueryEntity
}

type injector struct {
	dict    *dictionary
	dialect types.Dialect
	/*
		use check permission
		if user has permission to access this table and field
	*/
	fields     refFields
	textParams []string

	args       arguments
	numoFTable int
}

/*
Create a new dictionary
*/
func newDictionary() *dictionary {
	return &dictionary{
		fields:        dictionaryFields{},
		tableAlias:    make(map[string]string),
		entities:      make(map[string]*entity.Entity),
		aliasToEntity: make(map[string]*entity.Entity),
	}
}
func newInjector(dialect types.Dialect, textParam []string) *injector {
	return &injector{
		dict:       newDictionary(),
		dialect:    dialect,
		fields:     refFields{},
		textParams: textParam,
		//dynamicArgs: dynamicArgs,
		args: arguments{},
	}

}

const GET_PARAMS_FUNC = "dx__GetParams"

type CompilerError struct {
	Message string
	Args    []any

	Type      ERR_TYPE
	TraceArgs []any
}
type ERR_TYPE int

const (
	ERR_UNKNOWN ERR_TYPE = iota
	ERR_DATASET_NOT_FOUND
	ERR_FIELD_REQUIRE_ALIAS
	ERR_FIELD_NOT_FOUND
	ERR_AMBIGUOUS_FIELD_NAME
	ERR_EXPRESION_REQUIRE_ALIAS
	ERR_SYNTAX
)

func (e *CompilerError) Error() string {
	msg := []string{
		fmt.Sprintf(e.Message, e.Args...),
	}
	for _, traceArgs := range e.TraceArgs {
		msg = append(msg, fmt.Sprintf("Detail: %s", traceArgs))
	}
	return strings.Join(msg, "\n")
}
func newCompilerError(errType ERR_TYPE, message string, args ...any) error {
	return &CompilerError{
		Message: message,
		Args:    args,
		Type:    errType,
	}
}
func traceCompilerError(err error, args ...any) error {
	if ce, ok := err.(*CompilerError); ok {
		if ce.TraceArgs == nil {
			ce.TraceArgs = []any{}
		}
		ce.TraceArgs = append(ce.TraceArgs, args...)
		return ce
	}
	return err
}

type joinTableExprInjector struct {
	index int
}
type smartSqlParserArgs []any

func (s smartSqlParserArgs) String() string {
	bff, err := json.MarshalIndent(s, " ", "  ")
	if err != nil {
		return err.Error()
	}
	return string(bff)
}

type SmartSqlParser struct {
	Query              string
	Args               smartSqlParserArgs
	ScopeAccess        refFields
	OutputFields       outputFields
	AccessScope        accessScopes
	Hash256AccessScope string
}
