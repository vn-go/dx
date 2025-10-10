package compiler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type compilerFilterType struct {
}
type CompilerFilterTypeResult struct {
	Expr      string
	FieldExpr string
	Fields    map[string]types.OutputExpr
	//map field: agg func name
	FieldInFunc      map[string]types.OutputExpr
	IsConstant       bool
	Args             []interface{}
	HasAggregateFunc bool
}

//	func (c *CompilerFilterTypeResult) GetExpr() string {
//		return c.Expr
//	}
//
//	func (c *CompilerFilterTypeResult) GetFieldExpr() string {
//		return c.FieldExpr
//	}
//
//	func (c *CompilerFilterTypeResult) GetFields() map[string]string {
//		return c.Fields
//	}
type emptyParam struct {
	Index int
}

func union(a, b map[string]types.OutputExpr) map[string]types.OutputExpr {
	ret := map[string]types.OutputExpr{}
	for k, v := range a {
		if _, ok := ret[k]; !ok {
			ret[k] = v
		}
	}
	for k, v := range b {
		if _, ok := ret[k]; !ok {
			ret[k] = v
		}
	}
	return ret
}
func (cmp *compilerFilterType) Resolve(dialect types.Dialect, strFilter string, fields map[string]types.OutputExpr, n sqlparser.SQLNode, args *[]any) (*CompilerFilterTypeResult, error) {
	// Use switch-case for clean handling of different SQL node types
	switch x := n.(type) {

	// --- 1. Comparison Expression (e.g., =, >, <, LIKE) ---
	case *sqlparser.ComparisonExpr:
		// RULE CHECK: The left side of the comparison must not be a constant value (SQLVal).
		if _, ok := x.Left.(*sqlparser.SQLVal); ok {
			return nil, NewCompilerError(fmt.Sprintf("Invalid comparison '%s': The left side of operator '%s' must be a Field or Expression, not a constant value.", strFilter, x.Operator))
		}

		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right, args)
		if err != nil {
			return nil, err
		}

		// RULE CHECK: The right side of the comparison must be a Constant or Parameter.
		if !right.IsConstant {
			return nil, NewCompilerError(fmt.Sprintf("Invalid comparison '%s': The right side of operator '%s' must be a Constant or Parameter, not a Field or Expression.", strFilter, x.Operator))
		}

		// Prevent comparison of identical expressions (e.g., 'field = field')
		if left.Expr == right.Expr || left.FieldExpr == right.FieldExpr {
			return nil, NewCompilerError(fmt.Sprintf("Invalid comparison '%s': both sides of the comparison are identical.", strFilter))
		}

		expr := left.Expr + " " + x.Operator + " " + right.Expr
		fieldExpr := left.FieldExpr + " " + x.Operator + " " + right.FieldExpr

		// Collect fields

		fieldsSelected := internal.UnionMap(left.Fields, right.Fields)

		fieldInFunc := union(left.FieldInFunc, right.FieldInFunc)
		isAggFuncCall := left.HasAggregateFunc || right.HasAggregateFunc
		if isAggFuncCall {
			fmt.Println(fieldInFunc)
		}
		return &CompilerFilterTypeResult{
			Expr:             expr,
			FieldExpr:        fieldExpr,
			Fields:           fieldsSelected,
			HasAggregateFunc: isAggFuncCall,
			FieldInFunc:      fieldInFunc,
		}, nil

	// --- 2. Binary Expression (e.g., +, -, *, /) ---
	case *sqlparser.BinaryExpr:
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right, args)
		if err != nil {
			return nil, err
		}

		expr := left.Expr + " " + x.Operator + " " + right.Expr

		// Collect fields
		fieldsSelected := internal.UnionMap(left.Fields, right.Fields)

		return &CompilerFilterTypeResult{
			Expr:   expr,
			Fields: fieldsSelected,
		}, nil

	// --- 3. Column Name (Field or Boolean Literal) ---
	case *sqlparser.ColName:
		name := strings.ToLower(x.Name.String())

		// Handle boolean literals (e.g., 'true', 'false') as constants
		if name == "yes" || name == "no" || name == "true" || name == "false" {
			*args = append(*args, dialect.ToBool(x.Name.String()))
			return &CompilerFilterTypeResult{
				Expr:       "?",
				IsConstant: true,
			}, nil
		}

		// Check if the column is a valid field
		if v, ok := fields[name]; ok {
			return &CompilerFilterTypeResult{
				Expr:      v.Expr.ExprContent,
				FieldExpr: dialect.Quote(x.Name.String()),
				Fields: map[string]types.OutputExpr{
					name: {
						Expr: v.Expr,
					},
				},
			}, nil
		}

		// Unknown field error
		strFields := make([]string, 0, len(fields))
		for k := range fields {
			strFields = append(strFields, k)
		}
		return nil, newCompilerError(fmt.Sprintf("Unknown field '%s'. Valid fields are: [%s]. Problem near '%s'.",
			x.Name.String(), strings.Join(strFields, ", "), strFilter), ERR)

	// --- 4. SQL Value (Constant or Parameter) ---
	case *sqlparser.SQLVal:
		v := string(x.Val)

		// Handle parameters (e.g., :v1)
		if strings.HasPrefix(v, ":v") {
			// pIndex, err := strconv.ParseInt(v[2:], 32, 0)
			// if err != nil {
			// 	return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter))
			// }
			// *args = append(*args, emptyParam{
			// 	Index: int(pIndex),
			// })
			return &CompilerFilterTypeResult{Expr: "?", FieldExpr: "?", IsConstant: true}, nil
		}

		// Handle specific literal types
		if x.Type == sqlparser.StrVal || internal.Helper.IsString(v) {
			*args = append(*args, internal.Helper.TrimStringLiteral(v))
			return &CompilerFilterTypeResult{
				Expr:      "?",
				FieldExpr: "?",

				IsConstant: true,
			}, nil
		}
		if internal.Helper.IsBool(v) {
			*args = append(*args, internal.Helper.ToBool(v))
			return &CompilerFilterTypeResult{
				Expr:       "?",
				FieldExpr:  "?",
				IsConstant: true,
			}, nil
		}
		if internal.Helper.IsNumber(v) {
			fValue, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter))
			}
			*args = append(*args, fValue)
			return &CompilerFilterTypeResult{
				Expr:       "?",
				FieldExpr:  "?",
				IsConstant: true,
			}, nil
		}
		if internal.Helper.IsFloatNumber(v) {
			fValue, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, NewCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter))
			}
			*args = append(*args, fValue)
			return &CompilerFilterTypeResult{
				Expr:       "?",
				FieldExpr:  "?",
				IsConstant: true,
			}, nil
		}

		// Invalid value error
		return nil, NewCompilerError(fmt.Sprintf("Invalid literal value '%s' in expression '%s'. The value type is unrecognized.", v, strFilter))

	// --- 5. Logical AND Expression (AndExpr) ---
	case *sqlparser.AndExpr:
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right, args)
		if err != nil {
			return nil, err
		}

		// RULE CHECK: Both sides must be Field or Expression (not Constant).
		if left.IsConstant || right.IsConstant {
			return nil, NewCompilerError(fmt.Sprintf("Invalid logic '%s': Operator 'AND' requires Field or Expression (e.g., a comparison) on both sides, not a constant value.", strFilter))
		}

		// Collect fields
		fieldsSelected := internal.UnionMap(left.Fields, right.Fields)

		fieldInFunc := union(left.FieldInFunc, right.FieldInFunc)

		return &CompilerFilterTypeResult{
			Expr:             left.Expr + " AND " + right.Expr,
			FieldExpr:        left.FieldExpr + " AND " + right.FieldExpr,
			Fields:           fieldsSelected,
			HasAggregateFunc: left.HasAggregateFunc || right.HasAggregateFunc,
			FieldInFunc:      fieldInFunc,
		}, nil

	// --- 6. Logical OR Expression (OrExpr) ---
	case *sqlparser.OrExpr:
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right, args)
		if err != nil {
			return nil, err
		}

		// RULE CHECK: Both sides must be Field or Expression (not Constant).
		if left.IsConstant || right.IsConstant {
			return nil, NewCompilerError(fmt.Sprintf("Invalid logic '%s': Operator 'OR' requires Field or Expression (e.g., a comparison) on both sides, not a constant value.", strFilter))
		}

		// Collect fields
		fieldsSelected := internal.UnionMap(left.Fields, right.Fields)

		return &CompilerFilterTypeResult{
			Expr:             left.Expr + " OR " + right.Expr,
			FieldExpr:        left.FieldExpr + " OR " + right.FieldExpr,
			Fields:           fieldsSelected,
			HasAggregateFunc: left.HasAggregateFunc || right.HasAggregateFunc,
		}, nil

	// --- 7. Logical NOT Expression (NotExpr) ---
	case *sqlparser.NotExpr:
		left, err := cmp.Resolve(dialect, strFilter, fields, x.Expr, args)
		if err != nil {
			return nil, err
		}

		// RULE CHECK: The negated expression must be a Field or Expression (not Constant).
		if left.IsConstant {
			return nil, NewCompilerError(fmt.Sprintf("Invalid negation '%s': Operator 'NOT' must be applied to a Field or Expression, not a constant value.", strFilter))
		}

		return &CompilerFilterTypeResult{
			Expr:             "NOT " + left.Expr,
			FieldExpr:        "NOT " + left.FieldExpr,
			Fields:           left.Fields,
			HasAggregateFunc: left.HasAggregateFunc,
		}, nil

	// --- 8. Function Expression ---
	case *sqlparser.FuncExpr:
		return cmp.ResolveFunc(dialect, strFilter, fields, x, args)

	// --- 9. Aliased Expression ---
	case *sqlparser.AliasedExpr:
		return cmp.Resolve(dialect, strFilter, fields, x.Expr, args)
	case *sqlparser.ParenExpr:
		return cmp.Resolve(dialect, strFilter, fields, x.Expr, args)
	// --- 10. Default / Unimplemented Node Type ---
	default:
		if isDebugMode {
			panic(fmt.Sprintf("Unimplemented SQL node type: %T (see Resolve in compilerFilter.go)", n))
		}
		return nil, newCompilerError(fmt.Sprintf("Invalid expression structure near '%s'. The expression format is not supported.", strFilter), ERR)
	}
}
func (cmp *compilerFilterType) ResolveFunc(dialect types.Dialect, strFilter string, fields map[string]types.OutputExpr, x *sqlparser.FuncExpr, args *[]any) (*CompilerFilterTypeResult, error) {
	strArgs := []string{}
	if x.Name.Lowered() == "contains" {
		if len(x.Exprs) != 2 {
			return nil, newCompilerError(fmt.Sprintf("%s require 2 args. expression is '%s", x.Name.String(), strFilter), ERR)
		}
		fieldsSelected := map[string]types.OutputExpr{}
		for _, e := range x.Exprs {
			ex, err := cmp.Resolve(dialect, strFilter, fields, e, args)
			if err != nil {
				return nil, err
			}
			fieldsSelected = internal.UnionMap(fieldsSelected, fieldsSelected)

			strArgs = append(strArgs, ex.Expr)
		}
		dialectDelegateFunction := types.DialectDelegateFunction{
			FuncName:         "CONCAT",
			Args:             []string{"'%'", strArgs[1], "'%'"},
			HandledByDialect: false,
		}
		ret, err := dialect.SqlFunction(&dialectDelegateFunction)
		if err != nil {

			return nil, err
		}
		if dialectDelegateFunction.HandledByDialect {

			return &CompilerFilterTypeResult{
				Expr:   ret,
				Fields: fieldsSelected,
			}, nil
		}
		ret = strArgs[0] + " LIKE " + dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")"
		return &CompilerFilterTypeResult{
			Expr:   ret,
			Fields: fieldsSelected,
		}, nil
	}
	fieldsSelected := map[string]types.OutputExpr{}
	fieldInArgs := map[string]types.OutputExpr{}
	fieldInFunc := map[string]types.OutputExpr{}
	for _, e := range x.Exprs {

		ex, err := cmp.Resolve(dialect, strFilter, fields, e, args)
		//fieldInArgs=append(fieldInArgs, ex.Fields[])
		if err != nil {
			return nil, err
		}
		fieldsSelected = internal.UnionMap(fieldsSelected, ex.Fields)
		if ex.Fields != nil {
			for k, v := range ex.Fields {
				if _, ok := fieldsSelected[k]; !ok {
					fieldInArgs[k] = types.OutputExpr{
						SqlNode: e,
						Expr:    v.Expr,
					}
					fields[k] = types.OutputExpr{
						SqlNode: e,
						Expr:    v.Expr,
					}
				}
			}
		}
		fieldInFunc = union(fieldInFunc, ex.FieldInFunc)
		strArgs = append(strArgs, ex.Expr)
	}

	dialectDelegateFunction := types.DialectDelegateFunction{
		FuncName:         x.Name.String(),
		Args:             strArgs,
		HandledByDialect: false,
	}
	ret, err := dialect.SqlFunction(&dialectDelegateFunction) //<-- has a while list prevent harmful function
	if err != nil {

		return nil, err
	}
	if dialectDelegateFunction.HandledByDialect {
		//newField := map[string]types.OutputExpr{}
		for k, v := range fieldsSelected {
			v.IsInAggregateFunc = dialectDelegateFunction.IsAggregate
			fieldsSelected[k] = v

		}
		return &CompilerFilterTypeResult{
			Expr:             ret,
			Fields:           fieldsSelected,
			HasAggregateFunc: dialectDelegateFunction.IsAggregate,
			FieldInFunc:      fieldsSelected,
		}, nil
	}
	if x.Name.Lowered() == "concat" {
		newArgs := []string{}
		for _, x := range dialectDelegateFunction.Args {
			newArgs = append(newArgs, "COALESCE("+x+",'')")
		}
		ret := dialectDelegateFunction.FuncName + "(" + strings.Join(newArgs, ", ") + ")"
		return &CompilerFilterTypeResult{
			Expr:   ret,
			Fields: fieldsSelected,
		}, nil
	}
	ret = dialectDelegateFunction.FuncName + "(" + strings.Join(dialectDelegateFunction.Args, ", ") + ")"
	if dialectDelegateFunction.IsAggregate {
		fmt.Println(fieldInArgs)
	}
	for k, v := range fieldsSelected {
		v.IsInAggregateFunc = dialectDelegateFunction.IsAggregate
		fieldsSelected[k] = v

	}
	return &CompilerFilterTypeResult{
		Expr:        ret,
		Fields:      fieldsSelected,
		FieldInFunc: fieldInArgs,

		HasAggregateFunc: dialectDelegateFunction.IsAggregate,
	}, nil
}

var CompilerFilter = &compilerFilterType{}
