package quicky

import (
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/sqlparser"
)

type ArgInspect struct {
}
type ArgInspects []ArgInspect
type FieldInspect struct {
	Expression        string
	Alias             string
	Typ               sqlparser.ValType
	IsInAggregateFunc bool
}
type FieldInspects map[string]FieldInspect
type MapSqlNode map[string]sqlNode
type QueryItem struct {
	Content  string
	NextType string
	Next     *QueryItem
	/*
		type of present is QueryItem or MapSqlNode
	*/
	InspectData MapSqlNode
}
type clause struct {
}
type sqlNode struct {
	nodes  []any
	source string
}

type clauseItem struct {
	Key     string
	Content string
}
type SqlParser struct {
	Statement string
	Args      []any
}
type DictionanryItem struct {
	Content string
	DbType  sqlparser.ValType
	Alias   string
}
type Dictionanry struct {
	FieldMap        map[string]DictionanryItem
	AliasMap        map[string]string
	AliasMapReverse map[string]string
	Entities        map[string]*entity.Entity
}
type JoinNode struct {
	Node     sqlparser.SQLNode
	JoinType string
	Next     *JoinNode
}

func newSqlParser() *SqlParser {
	return &SqlParser{}

}

const GET_PARAMS_FUNC = "dx__GetParams"

type BuildOnJoinResult struct {
	LeftTable  string
	RightTable string
	On         string
	JoinType   string
}
