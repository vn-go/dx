package compiler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/entity"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) createDictionaryForDelete(tables []string, fields map[string]types.OutputExpr) *Dictionary {
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
			tableAlias[strings.ToLower(x)] = x
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
	// typeToAlias := map[reflect.Type]string{}
	// c := 1
	moreMapEntity := map[string]*entity.Entity{}
	for tbl, x := range mapEntities {
		if mAlias, ok := manualAlaisMap[tbl]; ok {
			newMap[tbl] = mAlias
			// typeToAlias[x.EntityType] = mAlias
			moreMapEntity[strings.ToLower(mAlias)] = x

		}
		//else {
		// 	alais, foud := typeToAlias[x.EntityType]
		// 	if !foud {
		// 		typeToAlias[x.EntityType] = fmt.Sprintf("T%d", c)
		// 		newMap[tbl] = alais
		// 		c++
		// 	} else {
		// 		newMap[tbl] = alais
		// 	}
		// }
	}
	for k, v := range moreMapEntity {
		mapEntities[k] = v
	}
	for tbl, x := range mapEntities {
		// alias := typeToAlias[x.EntityType]
		for _, col := range x.Cols {

			key := strings.ToLower(fmt.Sprintf("%s.%s", tbl, col.Field.Name))
			ret.Field[key] = cmp.dialect.Quote(col.Name)
			ret.StructField[key] = col.Field

		}
	}

	ret.TableAlias = newMap
	if len(fields) > 0 {
		for k, v := range fields {
			ret.Field[k] = v.Expr.ExprContent
		}
	}
	return ret
}

func newCompilerDelete(sql, dbDriver string, skipQuoteExpression bool, getReturnField bool) (*compiler, error) {
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

	if stmDelete, ok := stm.(*sqlparser.Delete); ok {

		tableList := tableExtractor.getTables(stmDelete.TableExprs, make(map[string]bool))
		if tableList != nil {
			ret.dict = ret.createDictionaryForDelete(tableList.tables, nil)
		}

		return ret, nil
	}

	return nil, fmt.Errorf("compiler not support %s, %s", originalSql, `compiler\compiler.go`)

}
func CompileDelete(sql, dbDriver string, getReturnField bool, bySqlSelect bool) (*SqlCompilerInfo, error) {
	cmp, err := newCompilerDelete(sql, dbDriver, false, getReturnField)
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

	ret := &SqlCompilerInfo{
		Info:            info,
		Dict:            cmp.dict,
		Args:            info.Args,
		ExtraTextParams: cmp.extraParams,
	}
	if bySqlSelect {
		ret.Dict.ExprAlias = map[string]types.OutputExpr{}
		for k, x := range ret.Info.OutputFields {
			ret.Dict.ExprAlias[k] = x

		}
	}

	return ret, nil

}
