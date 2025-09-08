package compiler

import (
	"fmt"
	"reflect"
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
	mapEntities := model.ModelRegister.GetMapEntities(tables)
	ret := &Dictionary{
		TableAlias:  map[string]string{},
		Field:       map[string]string{},
		StructField: map[string]reflect.StructField{},
		Tables:      tables,
	}
	i := 1
	for tbl, x := range mapEntities {
		aliasTable := fmt.Sprintf("T%d", i)
		ret.TableAlias[tbl] = aliasTable
		for _, col := range x.Cols {
			key := strings.ToLower(fmt.Sprintf("%s.%s", tbl, col.Field.Name))
			ret.Field[key] = cmp.dialect.Quote(aliasTable, col.Name)
			ret.StructField[key] = col.Field
		}
		i++
	}
	return ret
}

func newCompiler(sql, dbDriver string) (*compiler, error) {

	originalSql := sql
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

	if stmSelect, ok := stm.(*sqlparser.Select); ok {

		tableList := tabelExtractor.getTables(stmSelect.From, make(map[string]bool))
		ret.dict = ret.CreateDictionary(tableList)

	} else {
		return nil, fmt.Errorf("compiler not support %s", originalSql)
	}
	return ret, nil
}

func (cmp *compiler) getSqlInfo() (*types.SqlInfo, error) {

	stmSelect := cmp.node.(*sqlparser.Select)
	strSelect, err := cmp.resolveSelect(stmSelect.SelectExprs)
	if err != nil {
		return nil, err
	}
	ret := &types.SqlInfo{
		StrSelect: strSelect,
	}
	return ret, nil

}
func (cmp *compiler) resolveSelect(selectExprs sqlparser.SelectExprs) (string, error) {
	fields := []string{}
	for _, selectExpr := range selectExprs {
		if starExpr, ok := selectExpr.(*sqlparser.StarExpr); ok {
			if !starExpr.TableName.IsEmpty() {
				tblName := starExpr.TableName.Name.String()

				ent := model.ModelRegister.FindEntityByName(tblName)
				if ent != nil {
					if tableAlais, found := cmp.dict.TableAlias[strings.ToLower(tblName)]; found {
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
func Compile(sql, dbDriver string) (*types.SqlInfo, error) {
	cmp, err := newCompiler(sql, dbDriver)
	if err != nil {
		return nil, err
	}
	return cmp.getSqlInfo()
}
