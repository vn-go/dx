package shorttest

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
	"github.com/vn-go/dx/sqlparser"
)

type InspectStatement struct {
	sqlparser.Statement

	TextParams                  []string
	Dialect                     types.Dialect
	Scope                       *ColumnsScope
	ColummsInDictToColumnsScope map[string]string
	ColumnsScope                map[string]string
	ColumnsDict                 map[string]string
	ColumnsFieldMap             map[string]entity.ColumnDef
	ColumnsDictRevert           map[string]string
	Entities                    []*entity.Entity
	AliasEntityName             map[string]string
	AliasEntityNameRevert       map[string]string
	AliasTableMap               map[string]string
	CurrentIndexOfArg           IndexArgInfo
	ParamArgs                   []any
	Tables                      []string
	IndexOfDynamic              int
}
type InspectNode struct {
	sqlparser.SQLNode
	UnionType string
	Next      *InspectNode
}
type InspectInfo struct {
	Contex   string
	Args     []any
	TextArgs []string
}

func (fx *InspectStatement) GetFrom() (*InspectInfo, error) {
	selector := fx.Statement.(*sqlparser.Select)
	firstArgs := selector.SelectExprs
	firstArg := firstArgs[0]
	if aliasExpr, ok := firstArg.(*sqlparser.AliasedExpr); ok {
		if expr, ok := aliasExpr.Expr.(*sqlparser.BinaryExpr); ok {
			var unionNode InspectNode = fx.GetInspectNode(expr)
			return fx.GetSourceFromInspectNode(unionNode)
		}
		if expr, ok := aliasExpr.Expr.(*sqlparser.FuncExpr); ok {
			fmt.Println(expr)
		}
		fmt.Println(aliasExpr)
	}
	// if fx, ok := firstArg.(*sqlparser.AliasedTableExpr); ok {
	// 	fmt.Println(fx)
	// }
	fmt.Println(firstArg)
	// ret := sqlparser.TableExprs{}

	panic("not implemented")
}

func (fx *InspectStatement) GetSourceFromInspectNode(unionNode InspectNode) (*InspectInfo, error) {
	panic("unimplemented")
}

func (fx *InspectStatement) GetInspectNode(expr *sqlparser.BinaryExpr) InspectNode {
	panic("unimplemented")
}
func (fx InspectStatement) ToSqlInfo() (*types.SqlInfo, error) {
	args := &internal.SqlArgs{}
	ret := &types.SqlInfo{}
	selector := fx.Statement.(*sqlparser.Select)
	startIndexOfArgs := 0
	for _, expr := range selector.SelectExprs {
		if aliasExpr, ok := expr.(*sqlparser.AliasedExpr); ok {
			if exprFn, ok := aliasExpr.Expr.(*sqlparser.FuncExpr); ok {
				if strings.ToLower(exprFn.Name.String()) == "from" {
					joinInfo, err := fx.ExtractJoin(exprFn, args)
					if err != nil {
						return nil, err
					}
					ret.From = joinInfo.Content
					ret.Args.ArgJoin = append(ret.Args.ArgJoin, (*args)[startIndexOfArgs:]...)
					startIndexOfArgs = args.Len()
					continue

				}
				if strings.ToLower(exprFn.Name.String()) == "select" {
					selectInfo, err := fx.ExtractSelect(exprFn, args)
					if err != nil {
						return nil, err
					}
					ret.StrSelect = selectInfo.Content
					ret.Args.ArgJoin = append(ret.Args.ArgJoin, (*args)[startIndexOfArgs:]...)
					startIndexOfArgs = args.Len()
					continue

				}
				if strings.ToLower(exprFn.Name.String()) == "where" {
					whereInfo, err := fx.ExtractWhere(exprFn, args)
					if err != nil {
						return nil, err
					}
					ret.StrWhere = whereInfo.Content
					ret.Args.ArgWhere = append(ret.Args.ArgWhere, (*args)[startIndexOfArgs:]...)
					startIndexOfArgs = args.Len()
					continue

				}
				panic(fmt.Sprintf("not implement %s,InspectStatement.ToSqlInfo,'%s'", exprFn.Name, `test\short_test\sqlparser.Statement.go`))

			}
		}
	}
	return ret, nil
}

func (fx InspectStatement) ExtractSelect(exprFn *sqlparser.FuncExpr, args *internal.SqlArgs) (*ResolveInfo, error) {
	selectItems := []string{}
	for _, x := range exprFn.Exprs {
		r, err := fx.Resolve(x, C_TYPE_SELECT, args)
		if err != nil {
			return nil, err
		}
		selectItems = append(selectItems, r.Content)
	}
	return &ResolveInfo{
		Content: strings.Join(selectItems, ", "),
	}, nil
}

type InspectArg struct {
	DynamicIndex int
	StaticIndex  int
	TextIndex    int
}

type TableInfo struct {
	Content string
}
type JoinConditionInfo struct {
	Content string
}

func (fx *InspectStatement) ExtractJoin(exprFn *sqlparser.FuncExpr, args *internal.SqlArgs) (*ResolveInfo, error) {
	cmps := fx.ExtractComparisonExpr(exprFn.Exprs)
	fmt.Println(cmps)
	tables := fx.ExtractTableAlias(exprFn.Exprs)
	if len(tables) == 1 {
		fx.BuildDictByExprSource(tables)
		tblName := strings.ToLower(fx.Tables[0])
		return &ResolveInfo{
			Content: fx.AliasTableMap[tblName],
		}, nil
	}

	items, err := fx.MakeJoin(tables[0], tables[1], args)
	if err != nil {
		return nil, err
	}
	fmt.Println(items)
	return &ResolveInfo{}, nil

}

func (fx *InspectStatement) BuildDictByExprSource(tables []*sqlparser.AliasedExpr) {
	//soucreItem := []string{}
	for i, x := range tables {
		alias := fmt.Sprintf("T%d", i+1)
		if !x.As.IsEmpty() {
			alias = x.As.String()
		}
		if colName, ok := x.Expr.(*sqlparser.ColName); ok {
			tableName := colName.Name.String()
			fx.BuildDict(tableName, alias)

		}
	}

}

func (fx *InspectStatement) MakeJoin(left *sqlparser.AliasedExpr, right *sqlparser.AliasedExpr, args *internal.SqlArgs) ([]ResolveInfo, error) {
	l, err := fx.Resolve(left.Expr, C_TYPE_JOIN, args)
	if err != nil {
		return nil, err
	}
	r, err := fx.Resolve(left.Expr, C_TYPE_JOIN, args)
	if err != nil {
		return nil, err
	}
	return []ResolveInfo{*l, *r}, nil
}

func (fx *InspectStatement) ExtractTableAlias(exprs sqlparser.SelectExprs) []*sqlparser.AliasedExpr {
	ret := []*sqlparser.AliasedExpr{}
	for _, expr := range exprs {

		if aliasedExpr, ok := expr.(*sqlparser.AliasedExpr); ok {
			if _, ok := aliasedExpr.Expr.(*sqlparser.ComparisonExpr); !ok {
				ret = append(ret, aliasedExpr)
			}
		}

	}
	return ret
}

func (fx *InspectStatement) ExtractComparisonExpr(exprs sqlparser.SelectExprs) []*sqlparser.ComparisonExpr {
	ret := []*sqlparser.ComparisonExpr{}
	for _, expr := range exprs {

		if aliasedExpr, ok := expr.(*sqlparser.AliasedExpr); ok {
			if cmpExpr, ok := aliasedExpr.Expr.(*sqlparser.ComparisonExpr); ok {
				ret = append(ret, cmpExpr)
			}
		}

	}
	return ret
}

func (fx *InspectStatement) GetJoinInfo(conditionnalExpr *sqlparser.ComparisonExpr, args *internal.SqlArgs) (*ResolveInfo, error) {
	left, err := fx.Resolve(conditionnalExpr.Left, C_TYPE_JOIN, args)
	if err != nil {
		return nil, err
	}
	right, err := fx.Resolve(conditionnalExpr.Left, C_TYPE_JOIN, args)
	if err != nil {
		return nil, err
	}
	return &ResolveInfo{
		Content: fmt.Sprintf("%s %s %s", left.Content, conditionnalExpr.Operator, right.Content),
	}, nil
}

type C_TYPE int

const (
	C_TYPE_JOIN C_TYPE = iota
	C_TYPE_SELECT
	C_TYPE_WHERE
	C_TYPE_FUNC
)

type RESOVLE_TYPE int

const (
	ARG_DYNAMIC RESOVLE_TYPE = iota
	ARG_TEXT
	ARG_CONSTANT
)

type ARG_TYPE int

const (
	ARG_TYPE_DYNAMIC ARG_TYPE = iota
	ARG_TYPE_STATIC_TEXT
	ARG_TYPE_CONST
)

//	type ArgInspect struct {
//		Typ ARG_TYPE
//		Val any
//	}
//
// type ArgInspects []ArgInspect
type ResolveInfo struct {
	Content           string
	Typ               sqlparser.ValType
	IsInAggregateFunc bool
	IsColumn          bool
	Children          map[string]ResolveInfo
}

func (fx *InspectStatement) ExtractAliasedExpr(expr *sqlparser.AliasedExpr, alias string) (*TableInfo, error) {
	if !expr.As.IsEmpty() {
		alias = expr.As.String()
	}
	//*sqlparser.ColName
	switch x := expr.Expr.(type) {
	case *sqlparser.ColName:
		tableName := x.Name.String()
		fx.BuildDict(tableName, alias)
		return &TableInfo{
			Content: fx.Dialect.Quote(tableName, alias),
		}, nil

	}
	panic("unimplemented")
}
func (q *InspectStatement) BuildDict(tableName string, alias string) error {
	ent := model.ModelRegister.FindEntityByName(tableName)
	if ent == nil {
		return NewQueryTypeError("dataset not found: %s", tableName)
	}
	if q.ColumnsDict == nil {
		q.ColumnsDict = make(map[string]string)
	}
	if q.Entities == nil {
		q.Entities = []*entity.Entity{}
	}
	if q.ColumnsScope == nil && q.Scope != nil && !q.Scope.IsFull {
		q.ColumnsScope = q.Scope.GetColScope()
	}
	if q.ColummsInDictToColumnsScope == nil {
		q.ColummsInDictToColumnsScope = make(map[string]string)
	}
	if q.AliasEntityName == nil {
		q.AliasEntityName = map[string]string{}
	}
	if q.AliasTableMap == nil {
		q.AliasTableMap = map[string]string{}
	}
	if q.AliasEntityNameRevert == nil {
		q.AliasEntityNameRevert = map[string]string{}
	}
	q.AliasTableMap[strings.ToLower(tableName)] = fmt.Sprintf("%s %s", q.Dialect.Quote(tableName), q.Dialect.Quote(alias))
	q.AliasEntityName[strings.ToLower(alias)] = ent.EntityType.Name()
	q.AliasEntityNameRevert[strings.ToLower(ent.EntityType.Name())] = alias
	q.Entities = append(q.Entities, ent)
	if q.Tables == nil {
		q.Tables = []string{}
	}
	if q.ColumnsFieldMap == nil {
		q.ColumnsFieldMap = make(map[string]entity.ColumnDef)

	}
	q.Tables = append(q.Tables, tableName)
	for _, col := range ent.Cols {
		key := strings.ToLower(fmt.Sprintf("%s.%s", alias, col.Field.Name))
		key2 := strings.ToLower(fmt.Sprintf("%s.%s", ent.EntityType.Name(), col.Field.Name))
		dbFieldName := q.Dialect.Quote(ent.TableName, col.Name)
		q.ColumnsDict[key] = dbFieldName
		q.ColumnsDict[key2] = dbFieldName
		q.ColummsInDictToColumnsScope[key] = tableName
		q.ColummsInDictToColumnsScope[key2] = tableName
		q.ColumnsFieldMap[strings.ToLower(dbFieldName)] = col
		q.ColumnsFieldMap[key] = col
		q.ColumnsFieldMap[key2] = col

	}
	return nil

}
