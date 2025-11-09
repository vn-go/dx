package dx

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sql"
	"github.com/vn-go/dx/sqlparser"
)

type frontEndQueryToSqlResult struct {
	Query        string
	Args         []any
	RequireScope sql.ExtractInfoReqiureAcessScope
	Output       sql.ExtractInfoOutputField
}

func (fSql *frontEndQueryToSqlResult) String() string {
	ret := fSql.Query
	for i, arg := range fSql.Args {
		if reflect.TypeOf(arg) == reflect.TypeFor[string]() {
			strVal := arg.(string)
			strVal = strings.ReplaceAll(strVal, "'", "''")
			ret = strings.ReplaceAll(ret, fmt.Sprintf("@p%d", i+1), fmt.Sprintf("'%s'", strVal))
		} else {
			ret = strings.ReplaceAll(ret, fmt.Sprintf("@p%d", i+1), fmt.Sprintf("%v", arg))
		}

	}
	return ret
}

// dx.enduser.query.frontEndQuery.ToSql.go
func (f *frontEndQuery) ToSql() (*frontEndQueryToSqlResult, error) {

	var err error
	if f.err != nil {
		return nil, f.err
	}
	args := []any{}
	hasAggregate := false
	nonAggFields := []frontEndQueryResult{}

	if f.selector != "" {
		args = append(args, f.selectorsArgs...)
		f.selectorsField, err = f.selectResolve(f.selector, f.selectorsArgs)
		if err != nil {

			return nil, err
		}
		strSelect := []string{}
		//f.selectorsFieldMap = map[string]frontEndQueryResult{}

		for _, item := range f.selectorsField {
			f.selectorsFieldMap[strings.ToLower(item.alias)] = item
			strSelect = append(strSelect, item.content+" "+f.db.Dialect.Quote(item.alias))
			if item.isAggregate {
				hasAggregate = true
			} else {
				nonAggFields = append(nonAggFields, item)
			}
		}

		f.sqlInfo.SelectStatement.Selector = strings.Join(strSelect, ", ")

		if err != nil {
			return nil, err
		}

	}
	if f.filter != "" {
		args = append(args, f.filterArgs...)
		f.filterField, err = f.filterResolve(f.filter, f.filterArgs)
		if err != nil {

			return nil, err
		}

		for _, field := range f.filterField {

			if field.isAggregate {
				hasAggregate = true
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

		//filterArgs, err := f.filterArgsCompile.ToArray(f.filterArgs)
		if err != nil {
			return nil, err
		}
		//argsCompile = append(argsCompile, filterArgs...)
	}
	if hasAggregate {
		groupByItems := []string{}
		for _, item := range nonAggFields {
			groupByItems = append(groupByItems, item.content)
		}
		if f.sqlInfo.SelectStatement.GroupBy != "" {
			f.sqlInfo.SelectStatement.GroupBy += "," + strings.Join(groupByItems, ",")
		} else {
			f.sqlInfo.SelectStatement.GroupBy = strings.Join(groupByItems, ",")
		}
	}
	argsCompile, err := f.args.ToArray(args)
	if err != nil {
		return nil, err
	}
	sqlRet := f.db.Dialect.GetSelectStatement(f.sqlInfo.SelectStatement)

	return &frontEndQueryToSqlResult{
		Query:        sqlRet,
		Args:         argsCompile,
		RequireScope: f.sqlInfo.RequireAcessScope,
		Output:       f.OutptFields,
	}, nil

}

type initFrontEndQueryFilterResolve struct {
	val  []frontEndQueryResult
	err  error
	arg  sql.Args
	once sync.Once
}

var initFrontEndQueryFilterResolveMap sync.Map

func (f *frontEndQuery) filterResolve(filter string, args []any) ([]frontEndQueryResult, error) {
	a, _ := initFrontEndQueryFilterResolveMap.LoadOrStore(filter, &initFrontEndQueryFilterResolve{})
	i := a.(*initFrontEndQueryFilterResolve)
	i.once.Do(func() {
		i.val, i.err = f.filterResolveNoCache(filter, args)
		i.arg = f.args
	})
	f.args = i.arg
	return i.val, i.err
}
func (f *frontEndQuery) filterResolveNoCache(filter string, args []any) ([]frontEndQueryResult, error) {
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
}
type initFrontEndQuerySelectResolve struct {
	val    []frontEndQueryResult
	args   sql.Args
	fields sql.ExtractInfoOutputField
	err    error
	once   sync.Once
}

var initFrontEndQuerySelectResolveMap sync.Map

func (f *frontEndQuery) selectResolve(selector string, args []any) ([]frontEndQueryResult, error) {
	a, _ := initFrontEndQuerySelectResolveMap.LoadOrStore(selector, &initFrontEndQuerySelectResolve{})
	i := a.(*initFrontEndQuerySelectResolve)
	i.once.Do(func() {
		strSelect, err := internal.Helper.QuoteExpression2(selector)
		if err != nil {
			i.err = err
			return
		}

		selectStmt, err := sqlparser.Parse("select " + strSelect)
		if err != nil {
			i.err = sql.NewCompilerError(sql.ERR_SYNTAX, "syntax error : '%s'", selector)

		}
		i.val, i.err = f.resolveSelect(selectStmt.(*sqlparser.Select), args)
		i.args = f.args
		i.fields = f.OutptFields
		i.fields.OutputFields = f.OutptFields.NewOutputFields()
		i.fields.OutputFieldMap = map[string]sql.OutputField{}
		for _, field := range i.val {
			f := sql.OutputField{
				Name:         field.alias,
				FieldType:    field.fieldType,
				IsCalculated: field.isAggregate || field.isExpr,
				Expression:   field.content,
			}

			i.fields.OutputFields = append(i.fields.OutputFields, f)
			i.fields.OutputFieldMap[strings.ToLower(field.alias)] = f
		}
		f.OutptFields.Hash256 = f.OutptFields.OutputFields.ToHas256Key()

	})
	f.args = i.args
	f.OutptFields = i.fields
	return i.val, i.err
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

			isAggregate: left.isAggregate || right.isAggregate,
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

			isAggregate: left.isAggregate || right.isAggregate,
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

			isAggregate: left.isAggregate || right.isAggregate,
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

			isAggregate: r.isAggregate,
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

			isAggregate: left.isAggregate || right.isAggregate,
		}, nil
	case *sqlparser.SQLVal:
		switch expr.Type {
		case sqlparser.ValArg:
			indexOfArg, err := internal.Helper.ToIntFromBytes(expr.Val[2:])
			if err != nil {
				return nil, err
			}
			f.args.Add(nil, false, indexOfArg)
			ret := &frontEndQueryResult{
				content:         f.db.Dialect.ToParam(len(f.args), expr.Type),
				originalContent: "?",
				fieldType:       reflect.TypeFor[any](),
				isExpr:          true,
			}

			return ret, nil
		case sqlparser.IntVal:
			argVal, err := internal.Helper.ToIntFromBytes(expr.Val)
			if err != nil {
				return nil, err
			}
			f.args.Add(argVal, true, 0)
			ret := &frontEndQueryResult{
				content:         f.db.Dialect.ToParam(len(f.args), expr.Type),
				originalContent: "?",
				fieldType:       reflect.TypeFor[int64](),
				isExpr:          true,
			}

			return ret, nil
		case sqlparser.FloatVal:
			argVal, err := internal.Helper.ToFloatFromBytes(expr.Val)
			if err != nil {
				return nil, err
			}
			f.args.Add(argVal, true, 0)
			ret := &frontEndQueryResult{
				content:         f.db.Dialect.ToParam(len(f.args), expr.Type),
				originalContent: "?",
				fieldType:       reflect.TypeFor[float64](),
				isExpr:          true,
			}

			return ret, nil
		case sqlparser.BitVal:
			argVal := internal.Helper.ToBoolFromBytes(expr.Val)

			f.args.Add(argVal, true, 0)
			ret := &frontEndQueryResult{
				content:         f.db.Dialect.ToParam(len(f.args), expr.Type),
				originalContent: "?",
				fieldType:       reflect.TypeFor[any](),
				isExpr:          true,
			}

			return ret, nil
		case sqlparser.StrVal:
			f.args.Add(string(expr.Val), true, 0)
			ret := &frontEndQueryResult{
				content:         f.db.Dialect.ToParam(len(f.args), expr.Type),
				originalContent: "?",
				fieldType:       reflect.TypeFor[any](),
				isExpr:          true,
			}

			return ret, nil
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
		}, nil
	}
	funcContent := fmt.Sprintf("%s(%s)", expr.Name.String(), strings.Join(strArgs, ", "))
	return &frontEndQueryResult{
		content:         funcContent,
		isAggregate:     d.IsAggregate,
		originalContent: fmt.Sprintf("%s(%s)", expr.Name.String(), strings.Join(strArgsOriginal, ", ")),
		fieldType:       internal.Helper.CombineTypeByFunc(expr.Name.Lowered(), argsType),
		isExpr:          true,
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
