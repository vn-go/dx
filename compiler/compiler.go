package compiler

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

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
	TableAlias  map[string]string
	Field       map[string]string
	StructField map[string]reflect.StructField
	Tables      []string
}
type compiler struct {
	dict       *Dictionary
	sql        string
	node       sqlparser.SQLNode
	dialect    types.Dialect
	paramIndex int
}

func (cmp *compiler) CreateDictionary(tables []string) *Dictionary {
	tableAlias := map[string]string{}
	tblList := []string{}
	i := 1
	manualAlaisMap := map[string]string{}
	for _, x := range tables {
		items := strings.Split(x, "\n")
		if len(items) > 1 {
			manualAlaisMap[strings.ToLower(items[0])] = items[1]
			tblList = append(tblList, items[0])
		} else {
			tableAlias[strings.ToLower(x)] = fmt.Sprintf("T%d", i)
			tblList = append(tblList, x)
			i++
		}
	}
	mapEntities := model.ModelRegister.GetMapEntities(tblList)
	ret := &Dictionary{
		TableAlias:  map[string]string{},
		Field:       map[string]string{},
		StructField: map[string]reflect.StructField{},
		Tables:      tables,
	}
	ret.TableAlias = tableAlias
	// mapEntityTypes := map[reflect.Type]string{}
	// count := 1
	newMap := map[string]string{}
	//mapAlias := map[string]string{}
	typeToAlias := map[reflect.Type]string{}
	c := 1
	for tbl, x := range mapEntities {
		if mAlias, ok := manualAlaisMap[tbl]; ok {
			newMap[tbl] = mAlias
			typeToAlias[x.EntityType] = mAlias
		} else {
			if _, ok := typeToAlias[x.EntityType]; !ok {
				typeToAlias[x.EntityType] = fmt.Sprintf("T%d", c)
				newMap[tbl] = fmt.Sprintf("T%d", c)
				c++
			}
		}
	}
	for tbl, x := range mapEntities {
		alias := typeToAlias[x.EntityType]
		for _, col := range x.Cols {

			key := strings.ToLower(fmt.Sprintf("%s.%s", tbl, col.Field.Name))
			ret.Field[key] = cmp.dialect.Quote(alias, col.Name)
			ret.StructField[key] = col.Field

		}
	}

	ret.TableAlias = newMap
	return ret
}

func newCompiler(sql, dbDriver string, skipQuoteExpression bool) (*compiler, error) {

	originalSql := sql
	if !skipQuoteExpression {
		sql = internal.Helper.QuoteExpression(sql)
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

	if stmSelect, ok := stm.(*sqlparser.Select); ok {

		tableList := tabelExtractor.getTables(stmSelect, make(map[string]bool))

		ret.dict = ret.CreateDictionary(tableList)
		return ret, nil

	}
	if stmUnion, ok := stm.(*sqlparser.Union); ok {
		tableList := tabelExtractor.getTables(stmUnion.Left, make(map[string]bool))
		tableList = append(tableList, tabelExtractor.getTables(stmUnion.Right, make(map[string]bool))...)
		ret.dict = ret.CreateDictionary(tableList)
		return ret, nil
	}
	return nil, fmt.Errorf("compiler not support %s, %s", originalSql, `compiler\compiler.go`)

}
func (cmp *compiler) getSqlInfo() (*types.SqlInfo, error) {

	if stmSelect, ok := cmp.node.(*sqlparser.Select); ok {
		return cmp.getSqlInfoBySelect(stmSelect)
	}
	if stmUnion, ok := cmp.node.(*sqlparser.Union); ok {
		var ret *types.SqlInfo
		var err error
		if left, ok := stmUnion.Left.(*sqlparser.Select); ok {
			ret, err = cmp.getSqlInfoBySelect(left)
			if err != nil {
				return nil, err
			}
		} else {
			panic(fmt.Sprintf("compiler.getSqlInfo: not support %T", stmUnion.Left))
		}

		if right, ok := stmUnion.Right.(*sqlparser.Select); ok {
			var next *types.SqlInfo
			next, err := cmp.getSqlInfoBySelect(right)
			if err != nil {
				return nil, err
			} else {
				ret.UnionType = stmUnion.Type
				ret.UnionNext = next
			}
		} else {
			panic(fmt.Sprintf("compiler.getSqlInfo: not support %T", stmUnion.Left))
		}
		return ret, nil
	}

	panic(fmt.Sprintf("compiler.getSqlInfo: not support %T", cmp.node))
}
func (cmp *compiler) getSqlInfoBySelect(stmSelect *sqlparser.Select) (*types.SqlInfo, error) {

	strSelect, err := cmp.resolveSelect(stmSelect.SelectExprs)

	if err != nil {
		return nil, err
	}
	strFrom, err := cmp.resolveFrom(stmSelect.From)
	if err != nil {
		return nil, err
	}
	strWhere := ""
	if stmSelect.Where != nil {
		strWhere, err = cmp.resolveWhere(stmSelect.Where)
		if err != nil {
			return nil, err
		}
	}
	strOrderBy := ""
	if stmSelect.OrderBy != nil {
		strOrderBy, err = cmp.resolveOrderBy(stmSelect.OrderBy)
		if err != nil {
			return nil, err
		}
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
		strGroupBy, err = cmp.resolveGroupBy(stmSelect.GroupBy)
		if err != nil {
			return nil, err
		}

	}
	strHaving := ""
	if stmSelect.Having != nil {
		strHaving, err = cmp.resolveWhere(stmSelect.Having)
		if err != nil {
			return nil, err
		}

	}
	ret := &types.SqlInfo{
		StrSelect:  strSelect,
		From:       strFrom,
		StrWhere:   strWhere,
		StrOrder:   strOrderBy,
		Limit:      limit,
		Offset:     offset,
		StrGroupBy: strGroupBy,
		StrHaving:  strHaving,
	}
	return ret, nil

}

func (cmp *compiler) resolveGroupBy(group sqlparser.GroupBy) (string, error) {
	groupItems := []string{}
	for _, x := range group {
		str, err := cmp.resolve(x, C_WHERE)
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
		strLimit, err := cmp.resolve(limit.Rowcount, C_LIMIT)
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
		strOffset, err := cmp.resolve(limit.Offset, C_OFFSET)
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
func (cmp *compiler) resolveOrderBy(orderBy sqlparser.OrderBy) (string, error) {
	/*

	 */
	sortLst := []string{}
	for _, x := range orderBy {
		fx, err := cmp.resolve(x.Expr, C_ORDER)
		if err != nil {
			return "", err
		}
		sortLst = append(sortLst, fx+" "+x.Direction)
	}
	return strings.Join(sortLst, ","), nil

}
func (cmp *compiler) resolveFrom(node sqlparser.TableExprs) (string, error) {
	ret := []string{}
	for _, x := range node {

		strRet, err := cmp.resolve(x, C_JOIN)
		if err != nil {
			return "", err
		}
		ret = append(ret, strRet)

	}
	return strings.Join(ret, ","), nil
}
func (cmp *compiler) resolveSelect(selectExprs sqlparser.SelectExprs) (string, error) {
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
				for key, fieldStr := range cmp.dict.Field {
					exprField := fieldStr + " " + cmp.dialect.Quote(cmp.dict.StructField[key].Name)
					fields = append(fields, exprField)
				}
				return strings.Join(fields, ","), nil
			}

		} else {
			strExpr, err := cmp.resolve(selectExpr, C_SELECT)
			if err != nil {
				return "", err
			}
			fields = append(fields, strExpr)
		}

	}
	return strings.Join(fields, ","), nil
}

func (cmp *compiler) resolveWhere(node *sqlparser.Where) (string, error) {
	return cmp.resolve(node.Expr, C_WHERE)
}
func Compile(sql, dbDriver string) (*types.SqlInfo, error) {
	cmp, err := newCompiler(sql, dbDriver, false)
	if err != nil {
		return nil, err
	}
	return cmp.getSqlInfo()
}
func compileNoQuote(sql, dbDriver string) (*types.SqlInfo, error) {
	cmp, err := newCompiler(sql, dbDriver, true)
	if err != nil {
		return nil, err
	}
	return cmp.getSqlInfo()
}
func newBasicCompiler(sql, dbDriver string) (*compiler, error) {

	sql = internal.Helper.QuoteExpression(sql)

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
	tableList := tabelExtractor.getTables(node, make(map[string]bool))
	cmp.dict = cmp.CreateDictionary(tableList)
}
func CompileJoin(JoinExpr, dbDriver string) (string, error) {
	key := fmt.Sprintf("%s@%s", JoinExpr, dbDriver)
	return internal.OnceCall(key, func() (string, error) {
		//cmp, err := newCompiler("select * form "+JoinExpr, dbDriver)
		cmp, err := newBasicCompiler("select * form "+JoinExpr, dbDriver)
		if err != nil {
			return "", err
		}
		stmSelect := cmp.node.(*sqlparser.Select)
		cmp.initDict(stmSelect.From)
		return cmp.resolveFrom(stmSelect.From)
	})

}

func CompileSelect(Selecttor, dbDriver string) (string, error) {
	key := fmt.Sprintf("%s@%s", Selecttor, dbDriver)
	return internal.OnceCall(key, func() (string, error) {
		cmp, err := newBasicCompiler("select  "+Selecttor, dbDriver)
		if err != nil {
			return "", err
		}
		stmSelect := cmp.node.(*sqlparser.Select)
		cmp.initDict(stmSelect.SelectExprs)
		return cmp.resolveSelect(stmSelect.SelectExprs)
	})

}
func GetSql(sqlInfo *types.SqlInfo, dbDriver string) (*types.SqlParse, error) {

	retSql, err := factory.DialectFactory.Create("mysql").BuildSql(sqlInfo)
	if err != nil {
		return nil, err
	}

	sqlInfo, err = Compile(retSql.Sql, dbDriver)
	if err != nil {
		return nil, err
	}

	retSql2, err := factory.DialectFactory.Create(dbDriver).BuildSql(sqlInfo)
	if err != nil {
		return nil, err
	}
	retSql2.ArgIndex = retSql.ArgIndex
	return retSql2, nil

}
