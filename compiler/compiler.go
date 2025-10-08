package compiler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
	"github.com/vn-go/dx/sqlparser"
)

type COMPILER int

const (
	C_SELECT COMPILER = iota
	C_JOIN
	C_WHERE
	C_GROUP
	C_HAVING
	C_ORDER
	C_LIMIT
	C_OFFSET
	C_FUNC
	C_UPDATE
	C_EXPR
)

type Dictionary struct {
	TableAlias       map[string]string
	Field            map[string]string
	ExprAlias        map[string]types.OutputExpr
	StructField      map[string]reflect.StructField
	FieldAlaisToExpr map[string]string
	Tables           []string
	SourceUpdate     []string
}

//	type OutputExpr struct {
//		Expr      sqlparser.SQLNode
//		FieldName string
//	}
type compiler struct {
	returnField map[string]types.OutputExpr
	dict        *Dictionary
	sql         string
	node        sqlparser.SQLNode
	dialect     types.Dialect

	args           internal.CompilerArgs
	extraParams    []string
	isFromSubQuery bool
}
type initCreateDictionary struct {
	val  *Dictionary
	once sync.Once
}

var cacheCreateDictionary sync.Map

func (cmp *compiler) CreateDictionary(tables []string, fields map[string]types.OutputExpr) *Dictionary {
	key := reflect.TypeFor[compiler]().String() + "/" + reflect.TypeFor[compiler]().PkgPath() + "://CreateDictionary" + strings.Join(tables, ",")

	actually, _ := cacheCreateDictionary.LoadOrStore(key, &initCreateDictionary{})
	init := actually.(*initCreateDictionary)
	init.once.Do(func() {
		init.val = cmp.createDictionary(tables, fields)
	})

	return init.val
}

type sqlCompilerError struct {
	err error
	sql string
}

func (sqlErr *sqlCompilerError) Error() string {
	return fmt.Sprintf("syntax error %s\n%s", sqlErr.err, sqlErr.sql)
}
func newCompiler(sql, dbDriver string, skipQuoteExpression bool, getReturnField bool) (*compiler, error) {
	var err error
	originalSql := sql
	strSql, textParams := internal.Helper.InspectStringParam(sql)
	if !skipQuoteExpression {
		sql, err = internal.Helper.QuoteExpression(strSql)
		if err != nil {
			return nil, &sqlCompilerError{
				err: err,
				sql: originalSql,
			}
		}
	}
	//sqlparser.Backtick("[]")

	stm, err := sqlparser.Parse(strSql)
	if err != nil {
		// args := &internal.SelectorTypesArgs{}
		return nil, &sqlCompilerError{
			err: err,
			sql: originalSql,
		}
	}

	ret := &compiler{
		sql:         originalSql,
		node:        stm,
		dialect:     factory.DialectFactory.Create(dbDriver),
		args:        internal.CompilerArgs{},
		extraParams: textParams,
	}

	if stmSelect, ok := stm.(*sqlparser.Select); ok {

		tableList, err := tableExtractor.getTablesFromSql(sql, stmSelect)

		if err != nil {
			return nil, err
		}
		ret.isFromSubQuery = tableList.isSubQuery
		var outputField map[string]types.OutputExpr = nil
		if getReturnField {
			ret.returnField, err = FieldExttractor.GetFieldAlais(stmSelect, map[string]bool{}, tableList.isSubQuery)
			outputField = ret.returnField
			if err != nil {
				return nil, err
			}
		}

		ret.dict = ret.CreateDictionary(tableList.tables, outputField)
		return ret, nil

	}
	if _, ok := stm.(*sqlparser.Union); ok {

		return ret, nil
	}
	if stmDelete, ok := stm.(*sqlparser.Delete); ok {

		tableList := tableExtractor.getTables(stmDelete.TableExprs, make(map[string]bool))
		if tableList != nil {
			ret.dict = ret.CreateDictionary(tableList.tables, nil)
		}

		return ret, nil
	}
	if stmUpdate, ok := stm.(*sqlparser.Update); ok {
		visited := make(map[string]bool)
		tableList := tableExtractor.getTables(stmUpdate, visited)

		ret.dict = ret.CreateDictionary(tableList.tables, nil)
		ret.dict.SourceUpdate = tableList.tables
		return ret, nil
	}
	return nil, fmt.Errorf("compiler not support %s, %s", originalSql, `compiler\compiler.go`)

}

func (cmp *compiler) getSqlInfo() (*types.SqlInfo, error) {

	if stmSelect, ok := cmp.node.(*sqlparser.Select); ok {
		ret, err := cmp.getSqlInfoBySelect(stmSelect)
		if err != nil {
			return nil, err
		}

		//ret.Args = cmp.args

		return ret, nil
	}
	if stmUnion, ok := cmp.node.(*sqlparser.Union); ok {
		return cmp.getSqlInfoFromUnion(stmUnion)
	}
	if stmDelete, ok := cmp.node.(*sqlparser.Delete); ok {
		return cmp.getSqlInfoByDelete(stmDelete)
	}
	if stmUpdate, ok := cmp.node.(*sqlparser.Update); ok {
		return cmp.getSqlInfoByUpdate(stmUpdate)
	}
	panic(fmt.Sprintf("compiler.getSqlInfo: not support %T", cmp.node))
}
func (cmp *compiler) getSqlInfoByUpdate(stmUpdate *sqlparser.Update) (*types.SqlInfo, error) {
	var err error
	ret := &types.SqlInfo{
		SqlType: types.SQL_UPDATE,
	}

	settors := []string{}
	argsUpdate := internal.SqlArgs{}
	for _, x := range stmUpdate.Exprs {
		strExpr, err := cmp.resolve(x, C_UPDATE, &argsUpdate)
		if err != nil {
			return nil, err
		}
		settors = append(settors, strExpr)

	}
	ret.From = cmp.dialect.Quote(cmp.dict.Tables[0])
	ret.StrSetter = strings.Join(settors, ",")
	if stmUpdate.Where != nil {
		ret.StrWhere, err = cmp.resolve(stmUpdate.Where, C_UPDATE, &argsUpdate)
		if err != nil {
			return nil, err
		}
		//alias := cmp.dict.TableAlias[strings.ToLower(cmp.dict.Tables[0])]
		ret.From = cmp.dialect.Quote(cmp.dict.Tables[0])
	}
	return ret, nil
}
func (cmp *compiler) getSqlInfoByDelete(stmDelete *sqlparser.Delete) (*types.SqlInfo, error) {
	var err error
	ret := &types.SqlInfo{
		SqlType: types.SQL_DELETE,
	}
	if stmDelete.Where != nil {
		ret.StrWhere, err = cmp.resolveWhere(stmDelete.Where, &cmp.args.ArgWhere)
		if err != nil {
			return nil, err
		}
		alias := cmp.dict.TableAlias[strings.ToLower(cmp.dict.Tables[0])]
		ret.From = cmp.dialect.Quote(cmp.dict.Tables[0]) + " " + cmp.dialect.Quote(alias)
	}

	return ret, nil
}
func (cmp *compiler) getSqlInfoBySelect(stmSelect *sqlparser.Select) (*types.SqlInfo, error) {
	// sqlArgs := internal.CompilerArgs{}
	// sqlArgs = internal.FillArrayToEmptyFields[internal.CompilerArgs, internal.SqlArgs](sqlArgs)
	strSelect, err := cmp.resolveSelect(stmSelect.SelectExprs, &cmp.args.ArgsSelect)

	if err != nil {
		return nil, err
	}
	// sqlArgs.ArgsSelect = append(sqlArgs.ArgsSelect, cmp.args.ArgsSelect...)
	strFrom, err := cmp.resolveFrom(stmSelect.From, &cmp.args.ArgJoin)
	if err != nil {
		return nil, err
	}
	// sqlArgs.ArgJoin = append(sqlArgs.ArgJoin, cmp.args.ArgJoin...)
	strWhere := ""
	if stmSelect.Where != nil {
		strWhere, err = cmp.resolveWhere(stmSelect.Where, &cmp.args.ArgWhere)
		if err != nil {
			return nil, err
		}
		// sqlArgs.ArgWhere = append(sqlArgs.ArgWhere, cmp.args.ArgWhere...)
	}
	strOrderBy := ""
	if stmSelect.OrderBy != nil {
		strOrderBy, err = cmp.resolveOrderBy(stmSelect.OrderBy, &cmp.args.ArgOrder)
		if err != nil {
			return nil, err
		}
		// sqlArgs.ArgOrder = append(sqlArgs.ArgOrder, cmp.args.ArgOrder...)
	}
	var limit, offset *uint64
	if stmSelect.Limit != nil {
		limit, offset, err = cmp.resolveLimit(stmSelect.Limit)
		if err != nil {
			return nil, err
		}
	}
	strGroupBy := ""
	if stmSelect.GroupBy != nil {
		strGroupBy, err = cmp.resolveGroupBy(stmSelect.GroupBy, &cmp.args.ArgGroup)
		if err != nil {
			return nil, err
		}
		// sqlArgs.ArgGroup = append(sqlArgs.ArgGroup, cmp.args.ArgGroup...)
	}
	strHaving := ""
	if stmSelect.Having != nil {
		strHaving, err = cmp.resolveWhere(stmSelect.Having, &cmp.args.ArgHaving)
		if err != nil {
			return nil, err
		}
		// sqlArgs.ArgHaving = append(sqlArgs.ArgHaving, cmp.args.ArgHaving...)
	}
	if cmp.returnField == nil {
		cmp.returnField = map[string]types.OutputExpr{}
	}
	ret := &types.SqlInfo{
		StrSelect:    strSelect,
		From:         strFrom,
		StrWhere:     strWhere,
		StrOrder:     strOrderBy,
		Limit:        limit,
		Offset:       offset,
		StrGroupBy:   strGroupBy,
		StrHaving:    strHaving,
		OutputFields: cmp.returnField,
		Args:         cmp.args,
	}
	return ret, nil

}

func (cmp *compiler) resolveGroupBy(group sqlparser.GroupBy, args *internal.SqlArgs) (string, error) {
	groupItems := []string{}
	for _, x := range group {
		str, err := cmp.resolve(x, C_WHERE, args)
		if err != nil {
			return "", err
		}
		groupItems = append(groupItems, str)
	}
	return strings.Join(groupItems, ","), nil

}
func (cmp *compiler) resolveLimit(limit *sqlparser.Limit) (*uint64, *uint64, error) {
	var retLimit, retOffset *uint64
	if limit.Rowcount != nil {
		strLimit, err := cmp.resolve(limit.Rowcount, C_LIMIT, &internal.SqlArgs{})
		if err != nil {
			return nil, nil, err
		}
		rLimit, err := strconv.ParseUint(strLimit, 10, 64) // (string, base, bitSize)
		if err != nil {
			return nil, nil, err
		}
		retLimit = &rLimit
	}
	if limit.Offset != nil {
		strOffset, err := cmp.resolve(limit.Offset, C_OFFSET, nil)
		if err != nil {
			return nil, nil, err
		}
		rOffset, err := strconv.ParseUint(strOffset, 10, 64) // (string, base, bitSize)
		if err != nil {
			return nil, nil, err
		}
		retOffset = &rOffset
	}
	return retLimit, retOffset, nil
}
func (cmp *compiler) resolveOrderBy(orderBy sqlparser.OrderBy, args *internal.SqlArgs) (string, error) {
	/*

	 */
	sortLst := []string{}
	for _, x := range orderBy {
		fx, err := cmp.resolve(x.Expr, C_ORDER, args)
		if err != nil {
			return "", err
		}
		sortLst = append(sortLst, fx+" "+x.Direction)
	}
	return strings.Join(sortLst, ","), nil

}
func (cmp *compiler) resolveFrom(node sqlparser.TableExprs, args *internal.SqlArgs) (string, error) {
	ret := []string{}
	for _, x := range node {

		strRet, err := cmp.resolve(x, C_JOIN, args)
		if err != nil {
			return "", err
		}
		ret = append(ret, strRet)

	}
	return strings.Join(ret, ","), nil
}
func (cmp *compiler) resolveSelect(selectExprs sqlparser.SelectExprs, args *internal.SqlArgs) (string, error) {
	fields := []string{}
	for _, selectExpr := range selectExprs {
		if starExpr, ok := selectExpr.(*sqlparser.StarExpr); ok {
			if !starExpr.TableName.IsEmpty() {
				tblName := starExpr.TableName.Name.String()

				ent := model.ModelRegister.FindEntityByName(tblName)
				if ent != nil {
					tableAlais := ""
					found := false
					tableAlais, found = cmp.dict.TableAlias[strings.ToLower(tblName)]
					if !found {
						tableAlais, found = cmp.dict.TableAlias[strings.ToLower(ent.TableName)]
					}
					if found {
						for _, c := range ent.Cols {
							fields = append(fields, cmp.dialect.Quote(tableAlais, c.Name)+" "+cmp.dialect.Quote(c.Field.Name))
						}
					} else {
						for _, c := range ent.Cols {
							//exprField := fmt.Sprintf("%s.%s %s", ent.TableName, c.Name, c.Field.Name)
							fields = append(fields, cmp.dialect.Quote(ent.TableName, c.Name)+" "+cmp.dialect.Quote(c.Field.Name))
						}
					}

				} else {
					return "", fmt.Errorf("ca not found Entity has table name %s", tblName)
				}

			} else {
				if len(cmp.dict.Field) > 0 {
					for key, fieldStr := range cmp.dict.Field {
						exprField := fieldStr + " " + cmp.dialect.Quote(cmp.dict.StructField[key].Name)
						fields = append(fields, exprField)
					}
					return strings.Join(fields, ","), nil
				} else {
					return "*", nil
				}
			}

		} else {
			strExpr, err := cmp.resolve(selectExpr, C_SELECT, args)
			if err != nil {
				return "", err
			}
			fields = append(fields, strExpr)
		}

	}
	return strings.Join(fields, ","), nil
}

func (cmp *compiler) resolveWhere(node *sqlparser.Where, args *internal.SqlArgs) (string, error) {
	return cmp.resolve(node.Expr, C_WHERE, args)
}

type SqlCompilerInfo struct {
	Info            *types.SqlInfo
	Dict            *Dictionary
	Args            internal.CompilerArgs
	ExtraTextParams []string
}

func Compile(sql, dbDriver string, getReturnField bool) (*SqlCompilerInfo, error) {
	return internal.OnceCall("compiler/"+dbDriver+"/"+sql, func() (*SqlCompilerInfo, error) {
		cmp, err := newCompiler(sql, dbDriver, false, getReturnField)
		if err != nil {
			return nil, err
		}

		info, err := cmp.getSqlInfo()

		if err != nil {
			return nil, err
		}
		info.SqlSource = sql
		if getReturnField {
			if len(cmp.returnField) > 0 {
				info.OutputFields = cmp.returnField
			} else {
				tabble := cmp.dict.Tables[0]
				ent := model.ModelRegister.FindEntityByName(tabble)
				if ent != nil {
					info.OutputFields = make(map[string]types.OutputExpr)
					for _, col := range ent.Cols {
						info.OutputFields[strings.ToLower(col.Field.Name)] = types.OutputExpr{
							SqlNode: &sqlparser.AliasedExpr{
								Expr: &sqlparser.ColName{
									Name: sqlparser.NewColIdent(col.Name),
								},
								As: sqlparser.NewColIdent(col.Field.Name),
							},
							FieldName: col.Name,
						}
					}
				}
			}

		}

		return &SqlCompilerInfo{
			Info:            info,
			Dict:            cmp.dict,
			Args:            info.Args,
			ExtraTextParams: cmp.extraParams,
		}, nil
	})

}

//	func compileNoQuote(sql, dbDriver string) (*types.SqlInfo, error) {
//		cmp, err := newCompiler(sql, dbDriver, true, false)
//		if err != nil {
//			return nil, err
//		}
//		return cmp.getSqlInfo()
//	}
func newBasicCompiler(sql, dbDriver string) (*compiler, error) {
	var err error
	sql, err = internal.Helper.QuoteExpression(sql)
	if err != nil {
		return nil, err
	}

	stm, err := sqlparser.Parse(sql)
	if err != nil {
		return nil, err
	}

	ret := &compiler{
		sql:     sql,
		node:    stm,
		dialect: factory.DialectFactory.Create(dbDriver),
	}

	return ret, nil
}
func (cmp *compiler) initDict(node sqlparser.SQLNode) {

	tableList := tableExtractor.getTables(node, make(map[string]bool))
	if tableList != nil {
		cmp.dict = cmp.CreateDictionary(tableList.tables, nil)
	}

}

type CompileJoinResult struct {
	Expr string
	Args internal.SqlArgs
}

func CompileJoin(JoinExpr, dbDriver string) (*CompileJoinResult, error) {
	key := fmt.Sprintf("%s@%s", JoinExpr, dbDriver)
	return internal.OnceCall(key, func() (*CompileJoinResult, error) {
		//cmp, err := newCompiler("select * form "+JoinExpr, dbDriver)
		cmp, err := newBasicCompiler("select * form "+JoinExpr, dbDriver)
		if err != nil {
			return nil, err
		}
		stmSelect := cmp.node.(*sqlparser.Select)
		cmp.initDict(stmSelect.From)
		args := internal.SqlArgs{}
		expr, err := cmp.resolveFrom(stmSelect.From, &args)
		if err != nil {
			return nil, err
		}
		return &CompileJoinResult{
			Expr: expr,
			Args: args,
		}, nil
	})

}

type initCompileSelect struct {
	val  *CompileSelectResult
	err  error
	once sync.Once
}

var cacheCompileSelect sync.Map

type CompileSelectResult struct {
	Exprs string
	Args  internal.SqlArgs
}

func CompileSelect(Selecttor, dbDriver string) (*CompileSelectResult, error) {
	key := fmt.Sprintf("%s@%s", Selecttor, dbDriver)
	actually, _ := cacheCompileSelect.LoadOrStore(key, &initCompileSelect{})
	init := actually.(*initCompileSelect)
	init.once.Do(func() {
		cmp, err := newBasicCompiler("select  "+Selecttor, dbDriver)
		if err != nil {
			init.err = err
			return
		}
		stmSelect := cmp.node.(*sqlparser.Select)
		cmp.initDict(stmSelect.SelectExprs)
		argsSelect := internal.SqlArgs{}
		expr, err := cmp.resolveSelect(stmSelect.SelectExprs, &argsSelect)
		if err != nil {
			init.err = err
			return
		}
		init.val = &CompileSelectResult{
			Exprs: expr,
			Args:  argsSelect,
		}

	})
	if init.err != nil {
		cacheCompileSelect.Delete(key)
		return nil, init.err
	}
	return init.val, init.err

}

func GetSql(sqlInfo *types.SqlInfo, dbDriver string) (*types.SqlParse, error) {

	return internal.OnceCall("compiler/GetSql"+sqlInfo.GetKey(), func() (*types.SqlParse, error) {
		retSql, err := factory.DialectFactory.Create("mysql").BuildSql(sqlInfo)
		if err != nil {
			return nil, err
		}

		info, err := Compile(retSql.Sql, dbDriver, true)
		if err != nil {
			return nil, err
		}
		sqlInfo = info.Info
		retSql2, err := factory.DialectFactory.Create(dbDriver).BuildSql(sqlInfo)
		if err != nil {
			return nil, err
		}
		retSql2.ArgIndex = retSql.ArgIndex
		return retSql2, nil
	})

}

// True panic if compiler error
// False return error in compiler error
var isDebugMode = true

func Mode(isDebug bool) {
	isDebugMode = isDebug
}
