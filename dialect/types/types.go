package types

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/migate/loader/types"
	"github.com/vn-go/dx/sqlparser"
)

// Certain functions are transformed during compilation based on the SQL dialect.
// The dialect decides whether a function needs to be adapted for the target database driver.
// If required, the function will be rewritten here.
type DialectDelegateFunction struct {
	FuncName string
	Args     []string
	// is aggregate function
	IsAggregate      bool
	HandledByDialect bool // ✅ Indicates if this function is allowed to be delegated to the dialect
}
type DIALECT_DB_ERROR_TYPE = int

const (
	DIALECT_DB_ERROR_TYPE_UNKNOWN DIALECT_DB_ERROR_TYPE = iota
	DIALECT_DB_ERROR_TYPE_DUPLICATE
	DIALECT_DB_ERROR_TYPE_REFERENCES // ✅ refrences_violation
	DIALECT_DB_ERROR_TYPE_REQUIRED
)

func ErrorMessage(t DIALECT_DB_ERROR_TYPE) string {
	switch t {
	case DIALECT_DB_ERROR_TYPE_UNKNOWN:
		return "unknown"
	case DIALECT_DB_ERROR_TYPE_DUPLICATE:
		return "duplicate"
	case DIALECT_DB_ERROR_TYPE_REFERENCES:
		return "references"
	case DIALECT_DB_ERROR_TYPE_REQUIRED:
		return "required"
	default:
		return "unknown"
	}
}
func (e DialectError) Error() string {
	return fmt.Sprintf("code=%s, %s: %s cols %v tables %v, entity fields %v", e.Code, ErrorMessage(e.ErrorType), e.ErrorMessage, strings.Join(e.DbCols, ","), e.Table, strings.Join(e.Fields, ","))
}
func (e DialectError) Unwrap() error {
	return e.Err
}

type DialectError struct {
	Err            error
	ErrorType      DIALECT_DB_ERROR_TYPE
	Code           string
	ErrorMessage   string
	DbCols         []string
	Fields         []string
	Table          string
	StructName     string
	RefTable       string   //<-- table cause error
	RefStructName  string   //<-- Struct cause error
	RefCols        []string //<-- Columns in database cause error
	RefFields      []string //<-- Fields in struct cause error
	ConstraintName string   //<-- Constraint name cause error
}

func (e *DialectError) Reload() {
	e.Code = "ERR" + fmt.Sprintf("%04d", e.ErrorType)
	e.ErrorMessage = ErrorMessage(e.ErrorType)
}

type SqlParse struct {
	Sql      string
	ArgIndex []reflect.StructField
}
type SqlFncType struct {
	Expr        string
	IsAggregate bool
}

type Dialect interface {
	// true debug mode, panic will raise instead returning error
	ReleaseMode(bool)

	LikeValue(val string) string

	ParseError(dbSchema *types.DbSchema, err error) error
	Name() string
	Quote(str ...string) string
	GetTableAndColumnsDictionary(db *sql.DB) (map[string]string, error)
	ToText(value string) string
	ToBool(value string) string
	ToParam(index int) string
	SqlFunction(delegator *DialectDelegateFunction) (string, error)
	MakeSqlInsert(tableName string, columns []entity.ColumnDef, data interface{}) (string, []interface{})
	NewDataBase(db *db.DB, dbName string) (string, error)
	LimitAndOffset(sql string, limit, offset *uint64, orderBy string) string
	/*
				Build sql info and produce index field of arg in SqlInfo

				Example:
					SqlInfo{
						StrSelect  "a,b,c"
		    			StrWhere   strin
					}

	*/
	BuildSql(info *SqlInfo) (*SqlParse, error)
}
type SqlInfoArgs struct {
	ArgWhere   reflect.StructField
	ArgsSelect reflect.StructField
	ArgJoin    reflect.StructField
	ArgGroup   reflect.StructField
	ArgHaving  reflect.StructField
	ArgOrder   reflect.StructField
	ArgSetter  reflect.StructField
}
type SQL_TYPE int

func (sqlType SQL_TYPE) String() string { // chữ S viết hoa
	switch sqlType {
	case SQL_SELECT:
		return "SELECT"
	case SQL_DELETE:
		return "DELETE"
	case SQL_UPDATE:
		return "UPDATE"
	case SQL_INSERT:
		return "INSERT"
	default:
		return "UNKNOWN"
	}
}

const (
	SQL_SELECT SQL_TYPE = iota
	SQL_DELETE
	SQL_UPDATE
	SQL_INSERT // ✅ refrences_violation

)

type OutputExpr struct {
	Expr      sqlparser.SQLNode
	FieldName string
}
type SqlInfo struct {
	OutputFields map[string]OutputExpr
	SqlType      SQL_TYPE
	FieldArs     SqlInfoArgs
	StrSelect    string

	StrWhere  string
	StrSetter string
	Limit     *uint64
	Offset    *uint64
	StrOrder  string

	From interface{} //<--string or SqlInfo

	StrGroupBy string

	StrHaving string

	UnionNext *SqlInfo
	UnionType string
}

func (info *SqlInfo) GetKey() string {
	ret := fmt.Sprintf("%s,%s/%s/%s/%s/%s",
		info.SqlType,
		info.StrSelect,
		info.StrWhere,
		info.StrOrder,
		info.StrHaving,
		info.StrGroupBy,
	)

	if info.Limit != nil {
		ret += fmt.Sprintf("/%d", *info.Limit)
	}
	if info.Offset != nil {
		ret += fmt.Sprintf("/%d", *info.Offset)
	}
	if strForm, ok := info.From.(string); ok {
		ret += "/" + strForm
	}
	if nextInfo, ok := info.From.(SqlInfo); ok {
		ret += "/" + nextInfo.GetKey()
	}
	if info.UnionNext != nil {
		ret += "/" + info.UnionNext.GetKey()
	}
	return ret
}

// Pool global
var sqlInfoPool = sync.Pool{
	New: func() interface{} {
		return &SqlInfo{}
	},
}

// Lấy SqlInfo từ pool
func GetSqlInfo() *SqlInfo {
	return sqlInfoPool.Get().(*SqlInfo)
}

// Trả SqlInfo về pool
func PutSqlInfo(s *SqlInfo) {
	if s == nil {
		return
	}

	// Reset map
	for k := range s.OutputFields {
		delete(s.OutputFields, k)
	}
	// Không cần set s.OutputFields = nil nếu muốn reuse map
	s.SqlType = 0
	s.FieldArs = SqlInfoArgs{}
	s.StrSelect = ""
	s.StrWhere = ""
	s.StrSetter = ""
	s.Limit = nil
	s.Offset = nil
	s.StrOrder = ""
	s.From = nil
	s.StrGroupBy = ""
	s.StrHaving = ""
	s.UnionNext = nil
	s.UnionType = ""

	sqlInfoPool.Put(s)
}

// Clone dùng pool
func (s *SqlInfo) Clone() *SqlInfo {
	if s == nil {
		return nil
	}

	clone := GetSqlInfo()

	// Reset lại các field (đảm bảo sạch)
	for k := range clone.OutputFields {
		delete(clone.OutputFields, k)
	}
	if clone.OutputFields == nil && s.OutputFields != nil {
		clone.OutputFields = make(map[string]OutputExpr, len(s.OutputFields))
	}

	// Copy các giá trị cơ bản
	clone.SqlType = s.SqlType
	clone.FieldArs = s.FieldArs
	clone.StrSelect = s.StrSelect
	clone.StrWhere = s.StrWhere
	clone.StrSetter = s.StrSetter
	clone.StrOrder = s.StrOrder
	clone.StrGroupBy = s.StrGroupBy
	clone.StrHaving = s.StrHaving
	clone.UnionType = s.UnionType

	// Copy Limit / Offset
	if s.Limit != nil {
		limit := *s.Limit
		clone.Limit = &limit
	} else {
		clone.Limit = nil
	}
	if s.Offset != nil {
		offset := *s.Offset
		clone.Offset = &offset
	} else {
		clone.Offset = nil
	}

	// Deep copy OutputFields
	for k, v := range s.OutputFields {
		clone.OutputFields[k] = v
	}

	// Copy From
	switch v := s.From.(type) {
	case string:
		clone.From = v
	case *SqlInfo:
		clone.From = v.Clone() // đệ quy dùng pool
	default:
		clone.From = nil
	}

	// Copy UnionNext
	if s.UnionNext != nil {
		clone.UnionNext = s.UnionNext.Clone()
	} else {
		clone.UnionNext = nil
	}

	return clone
}
