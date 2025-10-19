package shorttest

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type QueryTypeError struct {
	Message string
}

func (e *QueryTypeError) Error() string {
	return e.Message
}
func (e *QueryTypeError) Is(target error) bool {
	_, ok := target.(*QueryTypeError)
	return ok
}
func NewQueryTypeError(message string, args ...interface{}) *QueryTypeError {
	return &QueryTypeError{
		Message: fmt.Sprintf(message, args...),
	}
}

type ColumnsScope struct {
	Cols   []string
	IsFull bool
}
type QueryType struct {
}
type IndexArgInfo struct {
	ConstIndex      int
	StaticTextIndex int
	DynamicIndex    int
	ScanIndex       int
}
type QueryInfo struct {
	Node                        sqlparser.SQLNode
	TextParams                  []string
	Dialect                     types.Dialect
	Scope                       *ColumnsScope
	ColummsInDictToColumnsScope map[string]string
	ColumnsScope                map[string]string
	ColumnsDict                 map[string]string
	Entities                    []*entity.Entity
	AliasEntityName             map[string]string
	CurrentIndexOfArg           IndexArgInfo
	ParamArgs                   []any
	//ColumnsDictRevert map[string]string
	//schemaLoader migartorType.IMigratorLoader
}

func (scope *ColumnsScope) GetKey() string {
	if scope.IsFull {
		return "*"
	}
	return fmt.Sprintf("%s", scope.Cols)
}

type initGetColScope struct {
	val  map[string]string
	once sync.Once
}

var initGetColScopeCache sync.Map

func (scope *ColumnsScope) GetColScope() map[string]string {
	a, _ := initGetColScopeCache.LoadOrStore(scope.GetKey(), &initGetColScope{})
	i := a.(*initGetColScope)
	i.once.Do(func() {
		i.val = make(map[string]string)
		for _, col := range scope.Cols {
			i.val[col] = strings.Split(col, ".")[0]
		}
	})
	return i.val
}
func (scope *ColumnsScope) Check(colName string) bool {
	if scope.IsFull {
		return true
	}
	colScope := scope.GetColScope()
	if _, ok := colScope[colName]; ok {
		return true
	}
	return false
}
func (q *QueryType) ToSql2(scope *ColumnsScope, db *dx.DB, statement string, args ...interface{}) (*types.SqlParse, error) {
	if scope == nil {
		scope = &ColumnsScope{
			IsFull: true,
		}
	}
	sql, textParams := internal.Helper.InspectStringParam(statement)

	sql, err := internal.Helper.QuoteExpression2(sql)
	if err != nil {
		return nil, err
	}

	sqlNode, err := sqlparser.Parse("select " + sql)

	if err != nil {
		return nil, err
	}
	sqlNodeInspect := &InspectStatement{
		Statement:  sqlNode,
		TextParams: textParams,
		Dialect:    db.Dialect,
		Scope:      scope,
		ParamArgs:  args,
	}

	sqlRet, err := sqlNodeInspect.ToSqlInfo()
	if err != nil {
		return nil, err
	}

	ret, err := sqlNodeInspect.Dialect.BuildSql(sqlRet)
	if err != nil {
		return nil, err
	}
	return ret, nil

}
func (q *QueryType) ToSql(scope *ColumnsScope, db *dx.DB, statement string, args ...interface{}) (*types.SqlParse, error) {
	if scope == nil {
		scope = &ColumnsScope{
			IsFull: true,
		}
	}
	sql, textParams := internal.Helper.InspectStringParam(statement)

	sql, err := internal.Helper.QuoteExpression(sql)
	if err != nil {
		return nil, err
	}
	sqlNode, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}

	ret := &QueryInfo{
		Node:       sqlNode,
		TextParams: textParams,
		Dialect:    db.Dialect,
		Scope:      scope,
		ParamArgs:  args,
		//schemaLoader: m.GetLoader(),
	}

	return ret.ToSQl()

}

var Query = &QueryType{}
