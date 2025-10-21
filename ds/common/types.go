package common

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/helper"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type FieldScope struct {
	DatasetName string
	FieldName   string
}
type AccessScope struct {
	AccessColums []string
	IsFullAccess bool
}

// This function is used to check if the field is allowed to access.
//
// fieldSelect is  RBMS basktic of field select.
//
// Example: mssql: []
func (a AccessScope) IsNotAllowAccess(fieldScope FieldScope) bool {
	if a.IsFullAccess {
		return false
	}
	panic(fmt.Sprintf("just impelemeted for full access"))
}

type ArgScaner struct {
	Value        any
	IsConstant   bool
	DynamicIndex int
}
type EXPR_TYPE int

const (
	// EXPR_TYPEFIELD means   database field in selector.
	EXPR_TYPE_FIELD EXPR_TYPE = iota
	// EXPR_TYPE_EXPR means   expression in selector(function or expression).
	EXPR_TYPE_EXPR
)

type Expression struct {
	Content         string
	OriginalContent string
	Type            EXPR_TYPE
	Alias           string
	// this expression is belong to function or not.
	//
	// Example: select(sum(a+b), a,c) -> a is belong to aggregate function, c is not belong to aggregate function.
	IsInAggregateFunc bool
}
type InjectInfo struct {
	Args        []ArgScaner
	DyanmicArgs []any
	TextArgs    []string
	Dialect     types.Dialect

	Dict                       *Dictionary
	SelectFields               map[string]Expression
	OuputFieldsInSelector      []string
	OuputFieldsInSelectorStack internal.Stack[[]string]
}
type DictionaryItem struct {
	Content string
	DbType  sqlparser.ValType
	Alias   string
}

type Dictionary struct {
	FieldMap map[string]DictionaryItem
	// map
	//
	// key: <dataset name in end-user query><field>
	//
	// value: FieldInfoDict {
	//
	// 	DatasetName: <dataset name in end-user query>
	//
	//  FieldName
	//
	//}
	// this info wil use in
	//
	// type common.AccessScope struct{}
	//
	// to check if the field is allowed to access.

	AliasMap        map[string]string
	AliasMapReverse map[string]string
	Entities        map[string]*entity.Entity
	/*
		Dung de tham chieu den tu table trong csdl den alias cua table o menh de join cua query

		Used to refer to a table in the database using the alias of the table in the join clause of the query.
	*/
	TableAlias map[string]string
}

func (d *Dictionary) ShowFieldMap() {
	if len(d.FieldMap) == 0 {
		fmt.Println("FieldMap is empty")
		return
	}
	for k, v := range d.FieldMap {
		fmt.Println(k, v.Content, v.DbType, v.Alias)
	}
}

func New(args []any, texts []string, dialect types.Dialect, scope AccessScope) *InjectInfo {
	injectInfo := &InjectInfo{
		Args:        []ArgScaner{},
		DyanmicArgs: args,
		TextArgs:    texts,
		Dialect:     dialect,

		Dict: NewDictionary(),
	}
	return injectInfo
}

type ResolverContent struct {
	// Content is the content to be resolved.
	// The end user should not see this content.
	Content string
	// OriginalContent is the original content after resolving. Use for error message.
	// The end user should see this content in error message.
	OriginalContent string
	AliasField      string
	DbTYpe          sqlparser.ValType
}
type ResolveQuery func(info *helper.InspectInfo, injectInfo *InjectInfo) (*types.SqlParse, error)
