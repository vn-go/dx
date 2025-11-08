package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sql"
	"github.com/vn-go/dx/sqlparser"
)

// dx.enduser.query.frontEndQuery.ToSql.go
func (f *frontEndQuery) ToSql() (string, []any, error) {
	var err error
	if f.err != nil {
		return "", nil, f.err
	}
	argsCompile := []any{}

	if f.selector != "" {
		f.selectorsField, err = f.selectResolve(f.selector, f.selectorsArgs)
		if err != nil {

			return "", nil, err
		}
		strSelect := []string{}
		//f.selectorsFieldMap = map[string]frontEndQueryResult{}
		for _, item := range f.selectorsField {
			f.selectorsFieldMap[strings.ToLower(item.alias)] = item
			strSelect = append(strSelect, item.content+" "+f.db.Dialect.Quote(item.alias))
			for _, arg := range item.args {

				if arg.indexArg > 0 {
					argsCompile = append(argsCompile, f.selectorsArgs[arg.indexArg])
				} else {
					argsCompile = append(argsCompile, arg.argVal)
				}
			}
		}

		f.selector = strings.Join(strSelect, ", ")
		f.selectorsArgs = argsCompile
	}
	if f.filter != "" {
		f.filterField, err = f.filterResolve(f.filter, f.filterArgs)
		if err != nil {

			return "", nil, err
		}
		for _, field := range f.filterField {
			if field.isAggregate {
				if f.sqlInfo.SelectStatement.Having != "" {
					f.sqlInfo.SelectStatement.Having += " AND (" + field.content + ")"
				} else {
					f.sqlInfo.SelectStatement.Having = field.content
				}
			} else {
				if f.sqlInfo.SelectStatement.Filter != "" {
					f.sqlInfo.SelectStatement.Filter += " AND (" + field.content + ")"
				} else {
					f.sqlInfo.SelectStatement.Filter = field.content
				}
			}
		}
	}
	sqlRet := f.db.Dialect.GetSelectStatement(f.sqlInfo.SelectStatement)
	return sqlRet, argsCompile, nil
}

func (f *frontEndQuery) filterResolve(filter string, args []any) ([]frontEndQueryResult, error) {
	ret := []frontEndQueryResult{}
	strSelect, err := internal.Helper.QuoteExpression2(filter)
	if err != nil {
		return nil, err
	}
	selectStmt, err := sqlparser.Parse("select " + strSelect)
	if err != nil {
		return nil, sql.NewCompilerError(sql.ERR_SYNTAX, "syntax error : '%s'. %s", err.Error(), filter)
	}

	nodes := f.filterOptimize(selectStmt.(*sqlparser.Select).SelectExprs[0].(*sqlparser.AliasedExpr).Expr)
	for _, node := range nodes {
		r, err := f.resoleExpr(node, args)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *r)
	}
	return ret, nil
}

func (f *frontEndQuery) filterOptimize(expr sqlparser.Expr) []sqlparser.SQLNode {
	switch expr := expr.(type) {
	case *sqlparser.AndExpr:
		return append(f.filterOptimize(expr.Left), f.filterOptimize(expr.Right)...)

	default:
		return []sqlparser.SQLNode{expr}
	}
}

type frontEndQueryResultArgs struct {
	indexArg int
	argVal   any
}
type frontEndQueryResult struct {
	alias           string
	content         string
	originalContent string
	isAggregate     bool
	isExpr          bool
	fieldType       reflect.Type

	args []frontEndQueryResultArgs
}

func (f *frontEndQuery) selectResolve(selector string, args []any) ([]frontEndQueryResult, error) {
	strSelect, err := internal.Helper.QuoteExpression2(selector)
	if err != nil {
		return nil, err
	}

	selectStmt, err := sqlparser.Parse("select " + strSelect)
	if err != nil {
		return nil, sql.NewCompilerError(sql.ERR_SYNTAX, "syntax error : '%s'", selector)
	}
	return f.resolveSelect(selectStmt.(*sqlparser.Select), args)
}

func (f *frontEndQuery) resolveSelect(selectStmt *sqlparser.Select, args []any) ([]frontEndQueryResult, error) {
	retItems := []frontEndQueryResult{}

	for _, expr := range selectStmt.SelectExprs {
		if expr, ok := expr.(*sqlparser.AliasedExpr); ok {
			ret, err := f.resoleExpr(expr.Expr, args)

			if err != nil {
				return nil, err
			}
			alias := expr.As.String()
			if expr.As.IsEmpty() {
				if ret.isAggregate || ret.isExpr {
					return nil, sql.NewCompilerError(sql.ERR_FIELD_REQUIRE_ALIAS, "'%s' require name(alias)", ret.originalContent)
				}
				alias = ret.alias
			}

			ret.alias = alias
			retItems = append(retItems, *ret)
			if f.selectorsFieldMap == nil {
				f.selectorsFieldMap = map[string]frontEndQueryResult{}
			}
			f.selectorsFieldMap[strings.ToLower(alias)] = *ret

		} else {
			return nil, sql.NewCompilerError(sql.ERR_SYNTAX, "syntax error : '%s'", "not support select expr")
		}
	}
	return retItems, nil
}

func (f *frontEndQuery) resoleExpr(expr sqlparser.SQLNode, args []any) (*frontEndQueryResult, error) {
	switch expr := expr.(type) {
	case *sqlparser.ColName:
		return f.resoleColName(expr, args)
	case *sqlparser.FuncExpr:
		return f.resoleFuncExpr(expr, args)
	case *sqlparser.AliasedExpr:
		return f.resoleExpr(expr.Expr, args)
	case *sqlparser.ComparisonExpr:
		left, err := f.resoleExpr(expr.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := f.resoleExpr(expr.Right, args)
		if err != nil {
			return nil, err
		}
		return &frontEndQueryResult{
			content:         fmt.Sprintf("%s %s %s", left.content, expr.Operator, right.content),
			originalContent: fmt.Sprintf("%s %s %s", left.originalContent, expr.Operator, right.originalContent),
			fieldType:       internal.Helper.CombineType(left.fieldType, right.fieldType, expr.Operator),
			isExpr:          true,
			args:            append(left.args, right.args...),
			isAggregate:     left.isAggregate || right.isAggregate,
		}, nil
	case *sqlparser.AndExpr:
		left, err := f.resoleExpr(expr.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := f.resoleExpr(expr.Right, args)
		if err != nil {
			return nil, err
		}
		return &frontEndQueryResult{
			content:         fmt.Sprintf("%s %s %s", left.content, "AND", right.content),
			originalContent: fmt.Sprintf("%s %s %s", left.originalContent, "and", right.originalContent),
			fieldType:       internal.Helper.CombineType(left.fieldType, right.fieldType, "and"),
			isExpr:          true,
			args:            append(left.args, right.args...),
			isAggregate:     left.isAggregate || right.isAggregate,
		}, nil
	case *sqlparser.OrExpr:
		left, err := f.resoleExpr(expr.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := f.resoleExpr(expr.Right, args)
		if err != nil {
			return nil, err
		}
		return &frontEndQueryResult{
			content:         fmt.Sprintf("%s %s %s", left.content, "OR", right.content),
			originalContent: fmt.Sprintf("%s %s %s", left.originalContent, "or", right.originalContent),
			fieldType:       internal.Helper.CombineType(left.fieldType, right.fieldType, "or"),
			isExpr:          true,
			args:            append(left.args, right.args...),
			isAggregate:     left.isAggregate || right.isAggregate,
		}, nil

	case *sqlparser.ParenExpr:
		r, err := f.resoleExpr(expr.Expr, args)
		if err != nil {
			return nil, err
		}
		return &frontEndQueryResult{
			content:         fmt.Sprintf("(%s)", r.content),
			originalContent: fmt.Sprintf("(%s)", r.originalContent),
			fieldType:       r.fieldType,
			isExpr:          true,
			args:            r.args,
			isAggregate:     r.isAggregate,
		}, nil
	case *sqlparser.BinaryExpr:

		left, err := f.resoleExpr(expr.Left, args)
		if err != nil {
			return nil, err
		}
		right, err := f.resoleExpr(expr.Right, args)
		if err != nil {
			return nil, err
		}
		return &frontEndQueryResult{
			content:         fmt.Sprintf("%s %s %s", left.content, expr.Operator, right.content),
			originalContent: fmt.Sprintf("%s %s %s", left.originalContent, expr.Operator, right.originalContent),
			fieldType:       internal.Helper.CombineType(left.fieldType, right.fieldType, expr.Operator),
			isExpr:          true,
			args:            append(left.args, right.args...),
			isAggregate:     left.isAggregate || right.isAggregate,
		}, nil
	case *sqlparser.SQLVal:
		switch expr.Type {
		case sqlparser.ValArg:
			indexOfArg, err := internal.Helper.ToIntFromBytes(expr.Val[2:])
			if err != nil {
				return nil, err
			}
			return &frontEndQueryResult{
				content:         "?",
				originalContent: "?",
				fieldType:       reflect.TypeFor[any](),
				isExpr:          true,
				args: []frontEndQueryResultArgs{
					{
						indexArg: indexOfArg,
					},
				},
			}, nil
		case sqlparser.IntVal:
			argVal, err := internal.Helper.ToIntFromBytes(expr.Val)
			if err != nil {
				return nil, err
			}
			return &frontEndQueryResult{
				content:         "?",
				originalContent: "?",
				fieldType:       reflect.TypeFor[int64](),
				isExpr:          true,
				args: []frontEndQueryResultArgs{
					{
						argVal: argVal,
					},
				},
			}, nil
		case sqlparser.FloatVal:
			argVal, err := internal.Helper.ToFloatFromBytes(expr.Val)
			if err != nil {
				return nil, err
			}
			return &frontEndQueryResult{
				content:         "?",
				originalContent: "?",
				fieldType:       reflect.TypeFor[float64](),
				isExpr:          true,
				args: []frontEndQueryResultArgs{
					{
						argVal: argVal,
					},
				},
			}, nil
		case sqlparser.BitVal:
			argVal := internal.Helper.ToBoolFromBytes(expr.Val)

			return &frontEndQueryResult{
				content:         "?",
				originalContent: "?",
				fieldType:       reflect.TypeFor[any](),
				isExpr:          true,
				args: []frontEndQueryResultArgs{
					{
						argVal: argVal,
					},
				},
			}, nil
		case sqlparser.StrVal:
			return &frontEndQueryResult{
				content:         "?",
				originalContent: "?",
				fieldType:       reflect.TypeFor[any](),
				isExpr:          true,
				args: []frontEndQueryResultArgs{
					{
						argVal: string(expr.Val),
					},
				},
			}, nil
		default:
			panic(fmt.Sprintf("unsupport sqlval type %T, ref frontEndQuery.resoleExpr", expr))
		}

	default:
		panic(fmt.Sprintf("unsupport expr type %T, ref frontEndQuery.resoleExpr", expr))
	}
}

func (f *frontEndQuery) resoleFuncExpr(expr *sqlparser.FuncExpr, args []any) (*frontEndQueryResult, error) {
	//arggExpr := []frontEndQueryResult{}
	strArgs := []string{}
	argsType := []reflect.Type{}
	strArgsOriginal := []string{}
	argRet := []frontEndQueryResultArgs{}
	for _, x := range expr.Exprs {
		exprStr, err := f.resoleExpr(x, args)
		if err != nil {
			return nil, err
		}
		if exprStr.isAggregate {
			return nil, sql.NewCompilerError(sql.ERR_SYNTAX, "You cannot use an aggregate function (like SUM, AVG, or COUNT) inside another")
		}
		//arggExpr = append(arggExpr, *exprStr)
		strArgs = append(strArgs, exprStr.content)
		strArgsOriginal = append(strArgsOriginal, exprStr.originalContent)
		argsType = append(argsType, exprStr.fieldType)
		argRet = append(argRet, exprStr.args...)
	}
	d := types.DialectDelegateFunction{
		FuncName: expr.Name.Lowered(),
		Args:     strArgs,
	}
	ret, err := f.db.Dialect.SqlFunction(&d)
	if err != nil {
		return nil, err
	}
	if d.HandledByDialect {
		return &frontEndQueryResult{
			content:         ret,
			isAggregate:     d.IsAggregate,
			originalContent: fmt.Sprintf("%s(%s)", expr.Name.String(), strings.Join(strArgsOriginal, ", ")),
			fieldType:       internal.Helper.CombineTypeByFunc(expr.Name.Lowered(), argsType),
			isExpr:          true,
			args:            argRet,
		}, nil
	}
	funcContent := fmt.Sprintf("%s(%s)", expr.Name.String(), strings.Join(strArgs, ", "))
	return &frontEndQueryResult{
		content:         funcContent,
		isAggregate:     d.IsAggregate,
		originalContent: fmt.Sprintf("%s(%s)", expr.Name.String(), strings.Join(strArgsOriginal, ", ")),
		fieldType:       internal.Helper.CombineTypeByFunc(expr.Name.Lowered(), argsType),
		isExpr:          true,
		args:            argRet,
	}, nil
}

func (fr *frontEndQuery) resoleColName(expr *sqlparser.ColName, args []any) (*frontEndQueryResult, error) {
	fieldName := expr.Name.Lowered()
	if f, ok := fr.sqlInfo.OuputInfo.OutputFieldMap[fieldName]; ok {
		return &frontEndQueryResult{
			content:         f.Expression,
			fieldType:       f.FieldType,
			originalContent: f.Name,
			alias:           f.Name,
		}, nil
	} else {
		if fr.selectorsFieldMap != nil {
			if f, ok := fr.selectorsFieldMap[strings.ToLower(fieldName)]; ok {
				return &frontEndQueryResult{
					content:         f.content,
					fieldType:       f.fieldType,
					originalContent: f.originalContent,
					alias:           f.alias,
					isExpr:          f.isExpr,
					isAggregate:     f.isAggregate,
				}, nil
			}
		}
		return nil, sql.NewCompilerError(sql.ERR_FIELD_NOT_FOUND, "field '%s' not found", fieldName)
	}

}
