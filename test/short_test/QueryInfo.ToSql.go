package shorttest

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
	"github.com/vn-go/dx/sqlparser"
)

type resolveType int

const (
	resolveTypeUnknown resolveType = iota
	resolveTypeSelect
	resolveTypeFilter
	resolveTypeFunc
	resolveTypeSource
	resolveJoinCondition
)

type EXPR_TYP int

const (
	EXPR_TYP_UNKNOWN EXPR_TYP = iota
	EXPR_TYP_COLUMN
	EXPR_TYP_PARAM

	EXPR_TYP_FUNC
	EXPR_TYP_EXPR
)

type ArgumentInfo struct {
	Val       any
	ArgType   ARG_TYPE
	RealIndex int
}
type expr struct {
	Content string
	// content before parse by sqlparser
	OriginalContent   string
	Args              []ArgumentInfo
	RefColumns        exprs
	IsInAggregateFunc bool
	ExprType          EXPR_TYP
	SqlNodType        sqlparser.ValType
}
type exprs []expr

func (e *exprs) Filter(fn func(expr *expr) bool) exprs {
	ret := exprs{}
	for _, x := range *e {
		if fn(&x) {
			ret = append(ret, x)
		}
	}
	return ret
}

func (e exprs) String() string {
	items := []string{}
	for _, x := range e {
		items = append(items, x.Content)
	}
	return strings.Join(items, ",")
}

func (q *QueryInfo) ToSQl() (*types.SqlParse, error) {
	selector, err := q.selectResolve(q.Node)
	if err != nil {
		return nil, err
	}
	fmt.Println(selector)
	sqlInfo := &types.SqlInfo{}
	return q.Dialect.BuildSql(sqlInfo)
}
func (q *QueryInfo) ExtractArgs(Exprs exprs) []any {
	ret := []any{}
	for _, fromExpr := range Exprs {
		for _, x := range fromExpr.Args {
			switch x.ArgType {
			case ARG_TYPE_CONST:
				ret = append(ret, x.Val)
			case ARG_TYPE_DYNAMIC:
				ret = append(ret, q.ParamArgs[x.RealIndex])
			case ARG_TYPE_STATIC_TEXT:
				ret = append(ret, q.TextParams[x.RealIndex])
			}

		}
	}
	return ret

}
func (q *QueryInfo) selectResolve(node sqlparser.SQLNode) (*types.SqlInfo, error) {
	ret := &types.SqlInfo{}
	if nodeSelect, ok := node.(*sqlparser.Select); ok {
		fromExprs, err := q.resolve(nodeSelect.From, resolveTypeSource)
		if err != nil {
			return nil, err
		}
		ret.From = fromExprs.String()
		ret.ArgumentData.ArgJoin = q.ExtractArgs(fromExprs)

		selectExpr, err := q.resolve(nodeSelect.SelectExprs, resolveTypeSelect)
		if err != nil {
			return nil, err
		}
		ret.StrSelect = selectExpr.String()
		ret.ArgumentData.ArgsSelect = q.ExtractArgs(selectExpr)
	}
	return ret, nil
}

func (q *QueryInfo) resolve(n sqlparser.SQLNode, resolver resolveType) (exprs, error) {
	switch n := n.(type) {
	case sqlparser.TableExprs:
		return q.resolveTableExprs(n, resolver)
	case *sqlparser.AliasedTableExpr:
		return q.resolveAliasedTableExpr(n, resolver)
	case sqlparser.SelectExprs:
		return q.resolveSelectExprs(n, resolver)

	}
	panic(fmt.Sprintf("not implemented for %T, in QueryInfo.resolve", n))
}

func (q *QueryInfo) resolveSelectExprs(n sqlparser.SelectExprs, resolver resolveType) (exprs, error) {
	ret := exprs{}
	for _, x := range n {
		r, err := q.resolveSelectExpr(x, resolver)
		if err != nil {
			return nil, err
		}
		if aliasExpr, ok := x.(*sqlparser.AliasedExpr); ok {
			if aliasExpr.As.IsEmpty() {
				return nil, NewQueryTypeError("'%s' expr must have alias", r.OriginalContent)
			}
		}
		ret = append(ret, *r)
	}

	return ret, nil
}

func (q *QueryInfo) resolveSelectExpr(n sqlparser.SelectExpr, resolver resolveType) (*expr, error) {
	switch n := n.(type) {
	case *sqlparser.StarExpr:
		return &expr{
			Content: "*",
			Args:    []ArgumentInfo{},
		}, nil
	case *sqlparser.AliasedExpr:
		return q.resolveAliasedExpr(n, resolver)
	}
	panic(fmt.Sprintf("not implemented for %T, in QueryInfo.resolveSelectExpr", n))
}

func (q *QueryInfo) resolveAliasedExpr(n *sqlparser.AliasedExpr, resolver resolveType) (*expr, error) {
	expr, err := q.resolveExpr(n.Expr, resolver)
	if err != nil {
		return nil, err
	}
	if !n.As.IsEmpty() {
		expr.Content = fmt.Sprintf("%s AS %s", expr.Content, q.Dialect.Quote(n.As.String()))
	}
	// if resolver == resolveTypeSelect && n.As.IsEmpty() && expr.ExprType != EXPR_TYP_COLUMN {
	// 	return nil, NewQueryTypeError("'%s' expr must have alias", expr.OriginalContent)
	// }
	return expr, nil
	//panic("unimplemented")
}

func (q *QueryInfo) resolveAliasedTableExpr(n *sqlparser.AliasedTableExpr, resolver resolveType) (exprs, error) {
	if tblNode, ok := n.Expr.(sqlparser.TableName); ok {
		fmt.Println(tblNode)
	}
	panic("unimplemented")
}

func (q *QueryInfo) resolveTableExprs(n sqlparser.TableExprs, resolver resolveType) (exprs, error) {
	ret := exprs{}

	for i, x := range n {
		alias := fmt.Sprintf("T%d", i+1)
		r, err := q.resolveTableExpr(x, alias, resolver)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *r)
	}

	return ret, nil
}

func (q *QueryInfo) resolveTableExpr(n sqlparser.TableExpr, alias string, resolver resolveType) (*expr, error) {
	switch n := n.(type) {
	case *sqlparser.AliasedTableExpr:

		if !n.As.IsEmpty() {
			alias = n.As.String()
		}
		if tblNode, ok := n.Expr.(sqlparser.TableName); ok {
			tableName := tblNode.Name.String()
			ent := model.ModelRegister.FindEntityByName(tableName)
			if ent == nil {
				return nil, NewQueryTypeError("dataset not found: %s", tableName)
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
			q.AliasEntityName[strings.ToLower(alias)] = ent.EntityType.Name()
			q.Entities = append(q.Entities, ent)

			for _, col := range ent.Cols {
				key := strings.ToLower(fmt.Sprintf("%s.%s", alias, col.Field.Name))
				key2 := strings.ToLower(fmt.Sprintf("%s.%s", ent.EntityType.Name(), col.Field.Name))
				dbFieldName := q.Dialect.Quote(ent.TableName, col.Name)
				q.ColumnsDict[key] = dbFieldName
				q.ColumnsDict[key2] = dbFieldName
				q.ColummsInDictToColumnsScope[key] = key2
				q.ColummsInDictToColumnsScope[key2] = key2

				//q.ColumnsDictRevert[dbFieldName] = key2
			}
			return &expr{
				Content: fmt.Sprintf("%s.%s", alias, ent.TableName),
				Args:    []ArgumentInfo{},
			}, nil

		}
	case *sqlparser.JoinTableExpr:
		left, err := q.resolveTableExpr(n.LeftExpr, alias, resolver)
		if err != nil {
			return nil, err
		}
		right, err := q.resolveTableExpr(n.RightExpr, alias, resolver)
		if err != nil {
			return nil, err
		}
		conditionalExpr, err := q.resolveJoinCondition(n.Condition, resolveJoinCondition)
		if err != nil {
			return nil, err
		}
		args := append(left.Args, right.Args...)
		args = append(args, (*conditionalExpr).Args...)
		return &expr{
			Content:         fmt.Sprintf("%s %s %s ON %s", left.Content, n.Join, right.Content, conditionalExpr.Content),
			OriginalContent: fmt.Sprintf("%s %s %s ON %s", left.OriginalContent, n.Join, right.OriginalContent, conditionalExpr.OriginalContent),
			Args:            args,
		}, nil
	}

	panic(fmt.Sprintf("not implemented for %T, in QueryInfo.resolveTableExpr", n))
}

func (q *QueryInfo) resolveJoinCondition(condition sqlparser.JoinCondition, resolveTypeFilter resolveType) (*expr, error) {
	return q.resolveExpr(condition.On, resolveTypeFilter)

}

func (q *QueryInfo) resolveExpr(n sqlparser.Expr, resolver resolveType) (*expr, error) {
	switch n := n.(type) {
	case *sqlparser.ComparisonExpr:
		left, err := q.resolveExpr(n.Left, resolver)
		if err != nil {
			return nil, err
		}
		right, err := q.resolveExpr(n.Right, resolver)
		if err != nil {
			return nil, err
		}
		refColums := exprs{*left, *right}
		refColums = refColums.Filter(func(x *expr) bool {
			return x.ExprType == EXPR_TYP_COLUMN
		})
		return &expr{
			Content:         fmt.Sprintf("%s %s %s", left.Content, n.Operator, right.Content),
			OriginalContent: fmt.Sprintf("%s %s %s", left.OriginalContent, n.Operator, right.OriginalContent),
			Args:            append(left.Args, right.Args...),
			ExprType:        EXPR_TYP_EXPR,
			RefColumns:      refColums,
		}, nil

	case *sqlparser.ColName:
		if resolver == resolveJoinCondition {
			return q.resolveJoinConditionColName(n, resolver)
		} else {
			return q.resolveColName(n, resolver)
		}
	case *sqlparser.FuncExpr:
		if n.Name.String() == internal.FnMarkSpecialTextArgs {
			//sqlparser.SelectExpr
			node := n.Exprs[0].(*sqlparser.AliasedExpr).Expr.(*sqlparser.SQLVal)

			indexOfStaticTextArg := -1
			indexOfStaticTextArg, err := internal.Helper.ToIntFormBytes(node.Val)
			if err != nil {
				return nil, err
			}
			defer func() {
				q.CurrentIndexOfArg.StaticTextIndex++
				q.CurrentIndexOfArg.ScanIndex++
			}()
			return &expr{
				Content:         q.Dialect.ToParam(q.CurrentIndexOfArg.ScanIndex, node.Type),
				OriginalContent: "'" + q.TextParams[indexOfStaticTextArg] + "'",
				Args: []ArgumentInfo{
					{
						Val:       q.TextParams[indexOfStaticTextArg],
						RealIndex: q.CurrentIndexOfArg.StaticTextIndex,
						ArgType:   ARG_TYPE_STATIC_TEXT,
					},
				},
				ExprType: EXPR_TYP_PARAM,
			}, nil
		}
		return q.resolveFuncExpr(n, resolver)
	case *sqlparser.BinaryExpr:
		return q.resolveBinaryExpr(n, resolver)
	case *sqlparser.SQLVal:
		return q.resolveSQLVal(n, resolver)
	case *sqlparser.AndExpr:
		return q.resolveAndExpr(n, resolver)
	case *sqlparser.OrExpr:
		return q.resolveOrExpr(n, resolver)
	case *sqlparser.NotExpr:
		return q.resolveNotExpr(n, resolver)
	}

	panic(fmt.Sprintf("not implemented for %T, in QueryInfo.resolveExpr", n))
}

func (q *QueryInfo) resolveNotExpr(n *sqlparser.NotExpr, resolver resolveType) (*expr, error) {
	r, err := q.resolveExpr(n.Expr, resolver)
	if err != nil {
		return nil, err
	}
	r.Content = fmt.Sprintf("NOT %s", r.Content)
	r.OriginalContent = fmt.Sprintf("NOT %s", r.OriginalContent)
	return r, nil
}

func (q *QueryInfo) resolveOrExpr(n *sqlparser.OrExpr, resolver resolveType) (*expr, error) {
	left, err := q.resolveExpr(n.Left, resolver)
	if err != nil {
		return nil, err
	}
	right, err := q.resolveExpr(n.Right, resolver)
	if err != nil {
		return nil, err
	}
	refColums := exprs{*left, *right}
	refColums = refColums.Filter(func(x *expr) bool {
		return x.ExprType == EXPR_TYP_COLUMN
	})
	return &expr{
		Content:         fmt.Sprintf("%s OR %s", left.Content, right.Content),
		OriginalContent: fmt.Sprintf("%s OR %s", left.OriginalContent, right.OriginalContent),
		Args:            append(left.Args, right.Args...),
		ExprType:        EXPR_TYP_EXPR,
		RefColumns:      refColums,
	}, nil
}

func (q *QueryInfo) resolveAndExpr(n *sqlparser.AndExpr, resolver resolveType) (*expr, error) {
	left, err := q.resolveExpr(n.Left, resolver)
	if err != nil {
		return nil, err
	}
	right, err := q.resolveExpr(n.Right, resolver)
	if err != nil {
		return nil, err
	}
	refColums := exprs{*left, *right}
	refColums = refColums.Filter(func(x *expr) bool {
		return x.ExprType == EXPR_TYP_COLUMN
	})
	return &expr{
		Content:         fmt.Sprintf("%s AND %s", left.Content, right.Content),
		OriginalContent: fmt.Sprintf("%s AND %s", left.OriginalContent, right.OriginalContent),
		Args:            append(left.Args, right.Args...),
		ExprType:        EXPR_TYP_EXPR,
		RefColumns:      refColums,
	}, nil
}

func (q *QueryInfo) resolveBinaryExpr(n *sqlparser.BinaryExpr, resolver resolveType) (*expr, error) {
	left, err := q.resolveExpr(n.Left, resolver)
	if err != nil {
		return nil, err
	}
	right, err := q.resolveExpr(n.Right, resolver)
	if err != nil {
		return nil, err
	}
	refColums := exprs{*left, *right}
	refColums = refColums.Filter(func(x *expr) bool {
		return x.ExprType == EXPR_TYP_COLUMN
	})
	return &expr{
		Content:         fmt.Sprintf("%s %s %s", left.Content, n.Operator, right.Content),
		OriginalContent: fmt.Sprintf("%s %s %s", left.OriginalContent, n.Operator, right.OriginalContent),
		Args:            append(left.Args, right.Args...),
		ExprType:        EXPR_TYP_EXPR,
		RefColumns:      refColums,
	}, nil
}

func (q *QueryInfo) resolveSQLVal(n *sqlparser.SQLVal, resolver resolveType) (*expr, error) {
	increseIndexOfArgs := func(t sqlparser.ValType) {

		q.CurrentIndexOfArg.ScanIndex++
		if t == sqlparser.ValArg {
			q.CurrentIndexOfArg.DynamicIndex++
		} else {
			q.CurrentIndexOfArg.ConstIndex++
		}
	}
	defer increseIndexOfArgs(n.Type)
	switch n.Type {
	case sqlparser.BitVal:

		val := internal.Helper.ToBoolFromBytes(n.Val)
		return &expr{
			Content:         q.Dialect.ToParam(q.CurrentIndexOfArg.ScanIndex, n.Type),
			OriginalContent: string(n.Val),
			Args: []ArgumentInfo{
				ArgumentInfo{
					Val:       val,
					ArgType:   ARG_TYPE_CONST,
					RealIndex: q.CurrentIndexOfArg.ConstIndex,
				},
			},
			ExprType:   EXPR_TYP_PARAM,
			SqlNodType: n.Type,
		}, nil
	case sqlparser.FloatVal:

		val, err := internal.Helper.ToFloatFormBytes(n.Val)
		if err != nil {
			return nil, err
		}
		return &expr{
			Content:         q.Dialect.ToParam(q.CurrentIndexOfArg.ScanIndex, n.Type),
			OriginalContent: string(n.Val),
			Args: []ArgumentInfo{
				ArgumentInfo{
					Val:       val,
					ArgType:   ARG_TYPE_CONST,
					RealIndex: q.CurrentIndexOfArg.ConstIndex,
				},
			},
			ExprType:   EXPR_TYP_PARAM,
			SqlNodType: n.Type,
		}, nil
	case sqlparser.IntVal:

		val, err := internal.Helper.ToIntFormBytes(n.Val)

		if err != nil {
			return nil, err
		}
		return &expr{
			Content:         q.Dialect.ToParam(q.CurrentIndexOfArg.ScanIndex, n.Type),
			OriginalContent: string(n.Val),
			Args: []ArgumentInfo{
				ArgumentInfo{
					Val:       val,
					ArgType:   ARG_TYPE_CONST,
					RealIndex: q.CurrentIndexOfArg.ConstIndex,
				},
			},
			ExprType:   EXPR_TYP_PARAM,
			SqlNodType: n.Type,
		}, nil
	case sqlparser.StrVal:

		val, err := internal.Helper.ToIntFormBytes(n.Val)
		if err != nil {
			return nil, err
		}
		return &expr{
			Content:         q.Dialect.ToParam(q.CurrentIndexOfArg.ScanIndex, n.Type),
			OriginalContent: "'" + string(n.Val) + "'",
			Args: []ArgumentInfo{
				ArgumentInfo{
					Val:       val,
					ArgType:   ARG_TYPE_CONST,
					RealIndex: q.CurrentIndexOfArg.ConstIndex,
				},
			},
			ExprType:   EXPR_TYP_PARAM,
			SqlNodType: n.Type,
		}, nil
	case sqlparser.ValArg:
		indexOfDynamicArgs, err := internal.Helper.ToIntFormBytes(n.Val[2:])
		if err != nil {
			return nil, err
		}

		return &expr{
			Content:         q.Dialect.ToParam(indexOfDynamicArgs, n.Type),
			OriginalContent: "'" + string(n.Val) + "'",
			Args: []ArgumentInfo{
				{
					RealIndex: indexOfDynamicArgs - 1, //q.CurrentIndexOfArg.DynamicIndex,
					ArgType:   ARG_TYPE_DYNAMIC,
				},
			},
			ExprType:   EXPR_TYP_PARAM,
			SqlNodType: n.Type,
		}, nil
	}
	panic(fmt.Sprintf("not implemented for %d, in QueryInfo.resolveSQLVal", n.Type))
}

func (q *QueryInfo) resolveFuncExpr(n *sqlparser.FuncExpr, resolver resolveType) (*expr, error) {
	//argsExprs := exprs{}
	refColumms := exprs{}
	delegator := &types.DialectDelegateFunction{
		FuncName: n.Name.String(),
		Args:     []string{},
	}
	ret := &expr{}
	originalArgs := []string{}
	for _, arg := range n.Exprs {
		argExpr, err := q.resolveSelectExpr(arg, resolver)
		if err != nil {
			return nil, err
		}
		originalArgs = append(originalArgs, argExpr.OriginalContent)
		refColumms = append(refColumms, (*argExpr).RefColumns...)
		refColumms = refColumms.Filter(func(x *expr) bool {
			return x.ExprType == EXPR_TYP_COLUMN
		})
		//argsExprs = append(argsExprs, *argExpr)
		delegator.Args = append(delegator.Args, argExpr.Content)
		delegator.ArgTypes = append(delegator.ArgTypes, argExpr.SqlNodType)
		ret.Args = append(ret.Args, argExpr.Args...)
	}

	content, err := q.Dialect.SqlFunction(delegator)
	if err != nil {
		return nil, err
	}
	if delegator.HandledByDialect {
		ret.Content = content
	} else {
		ret.Content = fmt.Sprintf("%s(%s)", n.Name.String(), strings.Join(delegator.Args, ","))
	}
	for i := range refColumms {
		refColumms[i].IsInAggregateFunc = delegator.IsAggregate
	}

	ret.ExprType = EXPR_TYP_FUNC
	ret.RefColumns = refColumms
	ret.IsInAggregateFunc = delegator.IsAggregate
	ret.OriginalContent = fmt.Sprintf("%s(%s)", n.Name.String(), strings.Join(originalArgs, ","))
	return ret, nil
}

func (q *QueryInfo) resolveColName(n *sqlparser.ColName, resolver resolveType) (*expr, error) {
	if n.Qualifier.IsEmpty() && len(q.Entities) > 1 {
		return nil, NewQueryTypeError("%s must have qualifier name in join condition", n.Name.String())
	}
	key := strings.ToLower(fmt.Sprintf("%s.%s", n.Qualifier.Name.String(), n.Name.String()))
	if q.ColumnsDict == nil {
		return nil, errors.New("ColumnsDict is nil")
	}

	if dbFieldName, ok := q.ColumnsDict[key]; ok {
		isOk := q.Scope.IsFull || q.Scope == nil
		if !isOk {
			keyLookupInScope, ok := q.ColummsInDictToColumnsScope[key]
			if !ok {
				entityName, ok := q.AliasEntityName[strings.ToLower(n.Qualifier.Name.String())]
				if !ok {
					return nil, NewQueryTypeError("Can not access to field '%s' in dataset '%s'", n.Name.String(), n.Qualifier.Name.String())
				}
				return nil, NewQueryTypeError("Can not access to field '%s' in dataset '%s'", n.Name.String(), entityName)
			}
			if _, found := q.ColumnsScope[keyLookupInScope]; found {
				isOk = found
			}

		}
		if !isOk {
			entityName, ok := q.AliasEntityName[strings.ToLower(n.Qualifier.Name.String())]
			if !ok {
				return nil, NewQueryTypeError("Can not access to field '%s' in dataset '%s'", n.Name.String(), n.Qualifier.Name.String())
			}
			return nil, NewQueryTypeError("Can not access to field '%s' in dataset '%s'", n.Name.String(), entityName)
		}
		originalContent := ""
		if n.Qualifier.IsEmpty() {
			originalContent = n.Name.String()
		} else {
			originalContent = fmt.Sprintf("%s.%s", n.Qualifier.Name.String(), n.Name.String())
		}
		return &expr{
			Content:         dbFieldName,
			ExprType:        EXPR_TYP_COLUMN,
			OriginalContent: originalContent,
		}, nil
	} else {
		return nil, NewQueryTypeError("field not found: %s", fmt.Sprintf("%s.%s", n.Qualifier.Name.String(), n.Name.String()))
	}

}

func (q *QueryInfo) resolveJoinConditionColName(n *sqlparser.ColName, resolver resolveType) (*expr, error) {
	if n.Qualifier.IsEmpty() {
		return nil, NewQueryTypeError("%s must have qualifier name in join condition", n.Name.String())
	}
	key := strings.ToLower(fmt.Sprintf("%s.%s", n.Qualifier.Name.String(), n.Name.String()))
	if q.ColumnsDict == nil {
		return nil, errors.New("ColumnsDict is nil")
	}
	if dbFieldName, ok := q.ColumnsDict[key]; ok {
		return &expr{
			Content:         dbFieldName,
			OriginalContent: fmt.Sprintf("%s.%s", n.Qualifier.Name.String(), n.Name.String()),
		}, nil
	} else {
		return nil, NewQueryTypeError("field not found: %s", fmt.Sprintf("%s.%s", n.Qualifier.Name.String(), n.Name.String()))
	}
}
