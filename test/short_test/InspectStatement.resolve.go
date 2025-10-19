package shorttest

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (fx *InspectStatement) Resolve(n sqlparser.SQLNode, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	switch t := n.(type) {
	case *sqlparser.ColName:
		return fx.ResolveColName(t, cType, args)

	case *sqlparser.FuncExpr:
		fnName := t.Name.Lowered()
		if fnName == "query" {
			return fx.ExtractSubQuery(t.Exprs)
		}
		return fx.ResolveFuncExpr(t, cType, args)
	case *sqlparser.AliasedExpr:
		return fx.ResolveAliasedExpr(t, cType, args)
	case *sqlparser.AndExpr:
		return fx.ResolveAndExpr(t, cType, args)
	case *sqlparser.OrExpr:
		return fx.ResolveOrExpr(t, cType, args)
	case *sqlparser.ComparisonExpr:
		return fx.ResolveComparisonExpr(t, cType, args)
	case *sqlparser.SQLVal:
		return fx.ResolveSQLVal(t, cType, args)
	case *sqlparser.BinaryExpr:
		return fx.ResolveBinaryExpr(t, cType, args)

	}
	panic(fmt.Sprintf("unimplemented: %T,in InspectStatement.Resolve, file '%s' ", n, `test\short_test\sqlparser.Statement.go`))
}



func (fx *InspectStatement) ResolveSQLVal(n *sqlparser.SQLVal, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	paramIndex := len(*args) + 1
	switch n.Type {
	case sqlparser.IntVal:
		val, err := internal.Helper.ToIntFormBytes(n.Val)
		if err != nil {
			return nil, err
		}
		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_CONSTANT,
			Value:     val,
		})
		return &ResolveInfo{
			Content: fx.Dialect.ToParam(paramIndex, n.Type),
		}, nil
	case sqlparser.BitVal:
		val := internal.Helper.ToBoolFromBytes(n.Val)

		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_CONSTANT,
			Value:     val,
		})
		return &ResolveInfo{
			Content: fx.Dialect.ToParam(paramIndex, n.Type),
		}, nil
	case sqlparser.FloatVal:
		val, err := internal.Helper.ToFloatFormBytes(n.Val)
		if err != nil {
			return nil, err
		}

		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_CONSTANT,
			Value:     val,
		})
		return &ResolveInfo{
			Content: fx.Dialect.ToParam(paramIndex, n.Type),
		}, nil
	case sqlparser.StrVal:
		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_CONSTANT,
			Value:     string(n.Val),
		})
		return &ResolveInfo{
			Content: fx.Dialect.ToParam(paramIndex, n.Type),
		}, nil
	case sqlparser.ValArg:

		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_DEFAULT,
			Value:     fx.ParamArgs[fx.IndexOfDynamic],
		})
		fx.IndexOfDynamic++
		return &ResolveInfo{
			Content: fx.Dialect.ToParam(paramIndex, n.Type),
		}, nil
	}
	panic(fmt.Sprintf("unimplemented: %T,in InspectStatement.ResolveSQLVal, file '%s' ", n, `test\short_test\sqlparser.Statement.go`))
}

func (fx *InspectStatement) ResolveFuncExpr(n *sqlparser.FuncExpr, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	if n.Name.String() == internal.FnMarkSpecialTextArgs {
		selectExpr := n.Exprs[0].(*sqlparser.AliasedExpr)
		sqlVal := selectExpr.Expr.(*sqlparser.SQLVal)
		paramIndex := len(*args) + 1
		apostropheIndex, err := internal.Helper.ToIntFormBytes(sqlVal.Val)
		if err != nil {
			return nil, err
		}
		*args = append(*args, internal.SqlArg{
			ParamType: internal.PARAM_TYPE_2APOSTROPHE,
			Value:     fx.TextParams[apostropheIndex],
		})
		return &ResolveInfo{
			Content: fx.Dialect.ToParam(paramIndex, sqlparser.StrVal),
		}, nil

	} else {
		oldType := cType
		defer func() {
			cType = oldType
		}()
		cType = C_TYPE_FUNC
		delegate := types.DialectDelegateFunction{
			FuncName: n.Name.String(),
		}
		children := map[string]ResolveInfo{}
		for _, x := range n.Exprs {
			r, err := fx.Resolve(x, cType, args)
			if err != nil {
				return nil, err
			}
			delegate.Args = append(delegate.Args, r.Content)
			delegate.ArgTypes = append(delegate.ArgTypes, r.Typ)
			if r.Children == nil && r.IsColumn {
				children = internal.UnionMap(children, map[string]ResolveInfo{
					r.Content: *r,
				})

			} else {
				children = internal.UnionMap(children, r.Children)
			}

		}

		content, err := fx.Dialect.SqlFunction(&delegate)
		if err != nil {
			return nil, err
		}
		if !delegate.HandledByDialect {
			return &ResolveInfo{
				Content:           delegate.FuncName + "(" + strings.Join(delegate.Args, ",") + ")",
				IsInAggregateFunc: delegate.IsAggregate,
				Children:          children,
			}, nil
		} else {
			return &ResolveInfo{
				Content:           content,
				IsInAggregateFunc: delegate.IsAggregate,
				Children:          children,
			}, nil
		}
	}

	panic(fmt.Sprintf("%s was not implemented in InspectStatement.ResolveFuncExpr, file '%s' ", n.Name.String(), `test\short_test\sqlparser.Statement.go`))
}

func (fx *InspectStatement) ResolveComparisonExpr(n *sqlparser.ComparisonExpr, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	children := map[string]ResolveInfo{}
	left, err := fx.Resolve(n.Left, cType, args)
	if err != nil {
		return nil, err
	}
	if left.IsColumn {
		children = internal.UnionMap(children, map[string]ResolveInfo{
			left.Content: *left,
		})
	} else {
		children = internal.UnionMap(children, left.Children)
	}
	right, err := fx.Resolve(n.Right, cType, args)
	if err != nil {
		return nil, err
	}
	if right.IsColumn {
		children = internal.UnionMap(children, map[string]ResolveInfo{
			right.Content: *right,
		})
	} else {
		children = internal.UnionMap(children, right.Children)
	}

	if err != nil {
		return nil, err
	}
	return &ResolveInfo{
		Content:           fmt.Sprintf("%s %s %s", left.Content, n.Operator, right.Content),
		Children:          children,
		IsInAggregateFunc: left.IsInAggregateFunc || right.IsInAggregateFunc,
		Typ:               sqlparser.BitVal,
	}, nil
}

func (fx *InspectStatement) ResolveOrExpr(n *sqlparser.OrExpr, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	children := map[string]ResolveInfo{}
	left, err := fx.Resolve(n.Left, cType, args)
	if err != nil {
		return nil, err
	}
	if left.IsColumn {
		children = internal.UnionMap(children, map[string]ResolveInfo{
			left.Content: *left,
		})
	} else {
		children = internal.UnionMap(children, left.Children)
	}
	right, err := fx.Resolve(n.Right, cType, args)
	if err != nil {
		return nil, err
	}
	if right.IsColumn {
		children = internal.UnionMap(children, right.Children)
	}

	return &ResolveInfo{
		Content:           fmt.Sprintf("%s OR %s", left.Content, right.Content),
		Typ:               sqlparser.BitVal,
		IsInAggregateFunc: left.IsInAggregateFunc || right.IsInAggregateFunc,
		Children:          children,
	}, nil
}

func (fx *InspectStatement) ResolveAndExpr(n *sqlparser.AndExpr, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	children := map[string]ResolveInfo{}
	left, err := fx.Resolve(n.Left, cType, args)
	if err != nil {
		return nil, err
	}
	if left.IsColumn {
		children = internal.UnionMap(children, map[string]ResolveInfo{
			left.Content: *left,
		})
	} else {
		children = internal.UnionMap(children, left.Children)
	}
	right, err := fx.Resolve(n.Right, cType, args)
	if err != nil {
		return nil, err
	}
	if right.IsColumn {
		children = internal.UnionMap(children, right.Children)
	}

	return &ResolveInfo{
		Content:           fmt.Sprintf("%s AND %s", left.Content, right.Content),
		Typ:               sqlparser.BitVal,
		IsInAggregateFunc: left.IsInAggregateFunc || right.IsInAggregateFunc,
		Children:          children,
	}, nil
}
func (fx *InspectStatement) ResolveBinaryExpr(n *sqlparser.BinaryExpr, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	children := map[string]ResolveInfo{}
	left, err := fx.Resolve(n.Left, cType, args)
	if err != nil {
		return nil, err
	}
	if left.IsColumn {
		children = internal.UnionMap(children, map[string]ResolveInfo{
			left.Content: *left,
		})
	} else {
		children = internal.UnionMap(children, left.Children)
	}
	right, err := fx.Resolve(n.Right, cType, args)
	if err != nil {
		return nil, err
	}
	if right.IsColumn {
		children = internal.UnionMap(children, right.Children)
	}

	return &ResolveInfo{
		Content:           fmt.Sprintf("%s %s %s", left.Content, n.Operator, right.Content),
		Typ:               right.Typ,
		IsInAggregateFunc: left.IsInAggregateFunc || right.IsInAggregateFunc,
		Children:          children,
	}, nil
}
func (fx *InspectStatement) ResolveAliasedExpr(n *sqlparser.AliasedExpr, cType C_TYPE, args *internal.SqlArgs) (*ResolveInfo, error) {
	r, err := fx.Resolve(n.Expr, cType, args)
	if err != nil {
		return nil, err
	}
	if cType == C_TYPE_SELECT {
		if n.As.IsEmpty() {
			colAlias := fx.ColumnsFieldMap[r.Content]
			r.Content += " " + fx.Dialect.Quote(colAlias.Field.Name)
			return r, nil
		} else {
			r.Content = fmt.Sprintf("%s AS %s", r.Content, fx.Dialect.Quote(n.As.String()))
		}

	}
	return r, nil
}
