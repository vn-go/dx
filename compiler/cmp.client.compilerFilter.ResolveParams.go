package compiler

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compilerFilterType) ResolveParams(dialect types.Dialect, strFilter string, fields map[string]types.OutputExpr, n *sqlparser.SQLVal, args *internal.SqlArgs, numberOfPreviuos2Apostrophe, startSqlIndex, startOdDynamicArg int) (*CompilerFilterTypeResult, error) {
	makeNewArgs := func(paramIndex int) int {
		indexOfParam := len(*args) + startSqlIndex + 1
		*args = append(*args, internal.SqlArg{
			ParamType:   internal.PARAM_TYPE_DEFAULT,
			IndexInSql:  indexOfParam,
			IndexInArgs: paramIndex - 1 + startOdDynamicArg,
		})
		return indexOfParam
	}
	makeNewConst := func(value any) int {
		indexInSql := len(*args) + startSqlIndex + 1
		*args = append(*args, internal.SqlArg{
			ParamType:  internal.PARAM_TYPE_CONSTANT,
			IndexInSql: indexInSql,
			Value:      value,
		})
		return indexInSql
	}
	switch n.Type {
	case sqlparser.ValArg:
		pIndex, err := internal.Helper.ToIntFormBytes(n.Val[2:])
		if err != nil {
			return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter))
		}
		indexOfParam := makeNewArgs(pIndex)
		return &CompilerFilterTypeResult{
			Expr:       dialect.ToParam(indexOfParam, n.Type),
			FieldExpr:  dialect.ToParam(indexOfParam, n.Type),
			IsConstant: true,
		}, nil

	case sqlparser.FloatVal:
		value, err := internal.Helper.ToFloatFormBytes(n.Val)
		if err != nil {
			return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter))
		}
		indexInSql := makeNewConst(value)
		return &CompilerFilterTypeResult{
			Expr:      dialect.ToParam(indexInSql, n.Type),
			FieldExpr: "?",

			IsConstant: true,
		}, nil
	case sqlparser.IntVal:
		value, err := internal.Helper.ToIntFormBytes(n.Val)
		if err != nil {
			return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter))
		}
		indexInSql := makeNewConst(value)
		return &CompilerFilterTypeResult{
			Expr:      dialect.ToParam(indexInSql, n.Type),
			FieldExpr: "?",

			IsConstant: true,
		}, nil
	case sqlparser.BitVal:
		value := internal.Helper.ToBoolFromBytes(n.Val)

		indexInSql := makeNewConst(value)
		return &CompilerFilterTypeResult{
			Expr:      dialect.ToParam(indexInSql, n.Type),
			FieldExpr: "?",

			IsConstant: true,
		}, nil
	default:
		indexInSql := makeNewConst(string(n.Val))
		return &CompilerFilterTypeResult{
			Expr:      dialect.ToParam(indexInSql, n.Type),
			FieldExpr: "?",

			IsConstant: true,
		}, nil

	}

	// // Handle parameters (e.g., :v1)
	// if strings.HasPrefix(v, ":v") {

	// }

	// // // Handle specific literal types
	// // if x.Type == sqlparser.StrVal || internal.Helper.IsString(v) {

	// // }
	// // if internal.Helper.IsBool(v) {
	// // 	indeInSql := len(*args) + startSqlIndex + 1
	// // 	*args = append(*args, internal.SqlArg{
	// // 		ParamType:  internal.PARAM_TYPE_CONSTANT,
	// // 		IndexInSql: indeInSql,
	// // 		Value:      internal.Helper.ToBool(v),
	// // 	})
	// // 	return &CompilerFilterTypeResult{
	// // 		Expr:       dialect.ToParam(indeInSql),
	// // 		FieldExpr:  "?",
	// // 		IsConstant: true,
	// // 	}, nil
	// // }
	// // if internal.Helper.IsNumber(v) {
	// // 	fValue, err := strconv.ParseInt(v, 10, 64)

	// // 	indexInSql := len(*args) + startSqlIndex + 1
	// // 	*args = append(*args, internal.SqlArg{
	// // 		ParamType:  internal.PARAM_TYPE_CONSTANT,
	// // 		IndexInSql: indexInSql,
	// // 		Value:      fValue,
	// // 	})
	// // 	return &CompilerFilterTypeResult{
	// // 		Expr:       dialect.ToParam(indexInSql),
	// // 		FieldExpr:  "?",
	// // 		IsConstant: true,
	// // 	}, nil
	// // }
	// // if internal.Helper.IsFloatNumber(v) {
	// // 	fValue, err := strconv.ParseFloat(v, 64)
	// // 	if err != nil {
	// // 		return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter))
	// // 	}
	// // 	indexInSql := len(*args) + startSqlIndex + 1
	// // 	*args = append(*args, internal.SqlArg{
	// // 		ParamType:  internal.PARAM_TYPE_CONSTANT,
	// // 		IndexInSql: indexInSql,
	// // 		Value:      fValue,
	// // 	})
	// // 	return &CompilerFilterTypeResult{
	// // 		Expr:       dialect.ToParam(indexInSql),
	// // 		FieldExpr:  "?",
	// // 		IsConstant: true,
	// // 	}, nil
	// }

	// //Invalid value error
	// return nil, NewCompilerError(fmt.Sprintf("Invalid literal value '%s' in expression '%s'. The value type is unrecognized.", v, strFilter))
}
