package compiler

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *cmpSelectorType) resolveValue(dialect types.Dialect, x *sqlparser.SQLVal, selector string, args *internal.SqlArgs, startOf2ApostropheArgs, startOfSqlIndex int) (*FieldSelect, error) {

	newArgs := func(index int) int {
		indexInSql := len(*args) + startOfSqlIndex + 1
		*args = append(*args, internal.SqlArg{
			ParamType:   internal.PARAM_TYPE_DEFAULT,
			IndexInSql:  indexInSql,
			IndexInArgs: index - 1,
		})
		return indexInSql
	}
	newConst := func(v any) int {
		indexInSql := len(*args) + startOfSqlIndex + 1
		*args = append(*args, internal.SqlArg{
			ParamType:  internal.PARAM_TYPE_CONSTANT,
			IndexInSql: indexInSql,
			Value:      v,
		})
		return indexInSql
	}

	switch x.Type {
	case sqlparser.ValArg:
		index, err := internal.Helper.ToIntFormBytes(x.Val[2:])
		if err != nil {
			return nil, NewCompilerError(fmt.Sprintf("%s is invalid ", selector))
		}
		indexInSql := newArgs(index)
		return &FieldSelect{
			Expr:         dialect.ToParam(indexInSql, x.Type),
			OriginalExpr: "?",
		}, nil
	case sqlparser.StrVal:
		value := string(x.Val)
		indexInSql := newConst(value)
		return &FieldSelect{
			Expr:         dialect.ToParam(indexInSql, x.Type),
			OriginalExpr: "'" + value + "'",
		}, nil
	case sqlparser.IntVal:
		value, err := internal.Helper.ToIntFormBytes(x.Val)
		if err != nil {
			return nil, NewCompilerError(fmt.Sprintf("%s is invalid ", selector))
		}
		indexInSql := newConst(value)
		return &FieldSelect{
			Expr:         dialect.ToParam(indexInSql, x.Type),
			OriginalExpr: string(x.Val),
		}, nil
	case sqlparser.FloatVal:
		value, err := internal.Helper.ToFloatFormBytes(x.Val)
		if err != nil {
			return nil, NewCompilerError(fmt.Sprintf("%s is invalid ", selector))
		}
		indexInSql := newConst(value)
		return &FieldSelect{
			Expr:         dialect.ToParam(indexInSql, x.Type),
			OriginalExpr: string(x.Val),
		}, nil

	case sqlparser.BitVal:
		value := internal.Helper.ToBoolFromBytes(x.Val)

		indexInSql := newConst(value)
		return &FieldSelect{
			Expr:         dialect.ToParam(indexInSql, x.Type),
			OriginalExpr: string(x.Val),
		}, nil
	default:
		value := string(x.Val)
		indexInSql := newConst(value)
		return &FieldSelect{
			Expr:         dialect.ToParam(indexInSql, x.Type),
			OriginalExpr: "'" + value + "'",
		}, nil
	}
	// v := string(x.Val)
	// if strings.HasPrefix(v, ":v") {
	// 	index, err := internal.Helper.ToInt(v[2:])
	// 	if err != nil {
	// 		return nil, NewCompilerError(fmt.Sprintf("%s is invalid ", selector))
	// 	}
	// 	indexInSql := len(*args) + startOfSqlIndex + 1
	// 	*args = append(*args, internal.SqlArg{
	// 		ParamType:   internal.PARAM_TYPE_DEFAULT,
	// 		IndexInSql:  indexInSql,
	// 		IndexInArgs: index - 1,
	// 	})
	// 	return &FieldSelect{
	// 		Expr:         dialect.ToParam(indexInSql),
	// 		OriginalExpr: "?",
	// 	}, nil

	// } else {
	// 	if x.Type == sqlparser.StrVal {
	// 		indexInSql := len(*args) + startOfSqlIndex + 1
	// 		*args = append(*args, internal.SqlArg{
	// 			ParamType:  internal.PARAM_TYPE_CONSTANT,
	// 			IndexInSql: indexInSql,
	// 			Value:      v,
	// 		})
	// 		return &FieldSelect{
	// 			Expr:         dialect.ToParam(indexInSql),
	// 			OriginalExpr: "'" + v + "'",
	// 		}, nil
	// 		//return dialect.ToText(v), nil
	// 	}
	// 	if internal.Helper.IsString(v) {
	// 		indexInSql := len(*args) + startOfSqlIndex + 1
	// 		*args = append(*args, internal.SqlArg{
	// 			ParamType:  internal.PARAM_TYPE_CONSTANT,
	// 			IndexInSql: indexInSql,
	// 			Value:      v,
	// 		})

	// 		return &FieldSelect{
	// 			Expr:         dialect.ToParam(indexInSql),
	// 			OriginalExpr: "'" + v + "'",
	// 		}, nil
	// 		//return dialect.ToText(v), nil
	// 	} else if internal.Helper.IsBool(v) {
	// 		indexInSql := len(*args) + startOfSqlIndex + 1
	// 		*args = append(*args, internal.SqlArg{
	// 			ParamType:  internal.PARAM_TYPE_CONSTANT,
	// 			IndexInSql: indexInSql,
	// 			Value:      internal.Helper.ToBool(v),
	// 		})

	// 		return &FieldSelect{
	// 			Expr:         dialect.ToParam(indexInSql),
	// 			OriginalExpr: v,
	// 		}, nil
	// 		//return dialect.ToBool(v), nil
	// 	} else if internal.Helper.IsFloatNumber(v) {
	// 		fx, err := internal.Helper.ToFloat(v)
	// 		if err != nil {
	// 			return nil, NewCompilerError(fmt.Sprintf("%s is invalid expression", selector))
	// 		}
	// 		indexInSql := len(*args) + startOfSqlIndex + 1
	// 		*args = append(*args, internal.SqlArg{
	// 			ParamType:  internal.PARAM_TYPE_CONSTANT,
	// 			IndexInSql: indexInSql,
	// 			Value:      fx,
	// 		})
	// 		return &FieldSelect{
	// 			Expr:         dialect.ToParam(indexInSql),
	// 			OriginalExpr: v,
	// 		}, nil
	// 		//return v, nil
	// 	} else if internal.Helper.IsNumber(v) {
	// 		fx, err := internal.Helper.ToInt(v)
	// 		if err != nil {
	// 			return nil, NewCompilerError(fmt.Sprintf("%s is invalid expression", cmp.originalSelector))
	// 		}
	// 		indexInSql := len(*args) + startOfSqlIndex + 1
	// 		*args = append(*args, internal.SqlArg{
	// 			ParamType:  internal.PARAM_TYPE_CONSTANT,
	// 			IndexInSql: indexInSql,
	// 			Value:      fx,
	// 		})
	// 		return &FieldSelect{
	// 			Expr:         dialect.ToParam(indexInSql),
	// 			OriginalExpr: v,
	// 		}, nil
	// 		//return v, nil
	// 	} else {
	// 		return nil, NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid value", v, cmp.originalSelector))
	// 	}

	// }
}
