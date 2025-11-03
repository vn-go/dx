package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// selectors.starExpr.go
func (s selectors) starExpr(expr *sqlparser.StarExpr, injector *injector) (*compilerResult, error) {
	strSelectItems := []string{}
	selectedExprs := dictionaryFields{}
	outputFields := []outputField{}
	selectedExprsReverse := dictionaryFields{}
	if expr.TableName.IsEmpty() {
		i := 1
		//get all tables in query
		tables := map[string]string{}
		keysTables := []string{}
		for k, v := range injector.dict.tableAlias {
			if _, ok := tables[v]; !ok {
				tables[v] = k
				keysTables = append(keysTables, v)
			}
		}
		keysTables = internal.Helper.SortStrings(keysTables)
		for _, qrAlias := range keysTables {
			aliasTable := fmt.Sprintf("T%d", i)

			if x, ok := injector.dict.entities[tables[qrAlias]]; ok {
				if fx, ok := injector.dict.tableAlias[strings.ToLower(x.EntityType.Name())]; ok {
					aliasTable = fx
				}
				for _, col := range x.Cols {
					aliasKey := col.Field.Name
					aliasField := injector.dialect.Quote(col.Field.Name)
					if len(injector.dict.entities) > 1 { // if there are more than one entity, we need to add entity name to alias
						aliasField = injector.dialect.Quote(x.EntityType.Name() + "_" + col.Field.Name)
						aliasKey = fmt.Sprintf("%s_%s", x.EntityType.Name(), col.Field.Name)
					}
					strSelectItems = append(strSelectItems, injector.dialect.Quote(aliasTable, col.Name)+" "+aliasField)
					selectedExprs[strings.ToLower(fmt.Sprintf("%s.%s", aliasTable, col.Field.Name))] = &dictionaryField{
						Expr:  injector.dialect.Quote(aliasTable, col.Name),
						Alias: col.Name,
					}
					refFieldKey := strings.ToLower(fmt.Sprintf("%s.%s", aliasTable, col.Field.Name))
					if _, ok := injector.fields[refFieldKey]; !ok {
						injector.fields[refFieldKey] = refFieldInfo{
							EntityName:      x.EntityType.Name(),
							EntityFieldName: col.Field.Name,
						}
					}

					outputFields = append(outputFields, outputField{
						Name:         aliasField,
						IsCalculated: false,
						FieldType:    internal.Helper.ToNullableType(col.Field.Type),
					})
					selectedExprsReverse[strings.ToLower(aliasKey)] = &dictionaryField{
						Expr:  injector.dialect.Quote(aliasTable, col.Name),
						Alias: col.Name,
					}
				}
			} else if subQuery, ok := injector.dict.subqueryEntites[qrAlias]; ok {
				for _, col := range subQuery.fields {
					strSelectItems = append(strSelectItems, injector.dialect.Quote(col.source, col.field)+" "+injector.dialect.Quote(col.field))
				}
			}

			i++
		}
	} else {
		if ent, ok := injector.dict.entities[strings.ToLower(expr.TableName.Name.String())]; ok {
			aliasTable := "T1"
			for _, col := range ent.Cols {
				strSelectItems = append(strSelectItems, injector.dialect.Quote(aliasTable, col.Name)+" "+injector.dialect.Quote(col.Field.Name))
				refFieldKey := strings.ToLower(fmt.Sprintf("%s.%s", ent.EntityType.Name(), col.Field.Name))
				if _, ok := injector.fields[refFieldKey]; !ok {
					injector.fields[refFieldKey] = refFieldInfo{
						EntityName:      ent.EntityType.Name(),
						EntityFieldName: col.Field.Name,
					}
				}
				selectedExprs[strings.ToLower(fmt.Sprintf("%s.%s", aliasTable, col.Field.Name))] = &dictionaryField{
					Expr:  injector.dialect.Quote(aliasTable, col.Name),
					Alias: col.Name,
				}
				selectedExprsReverse[strings.ToLower(col.Field.Name)] = &dictionaryField{
					Expr:  injector.dialect.Quote(aliasTable, col.Name),
					Alias: col.Name,
				}

			}
		} else {
			return nil, newCompilerError(ERR_DATASET_NOT_FOUND, "dataset %s not found", expr.TableName.Name.String())
		}
	}
	strSelect := strings.Join(strSelectItems, ", ")
	return &compilerResult{
		Content:              strSelect,
		Args:                 nil,
		Fields:               injector.fields,
		selectedExprs:        selectedExprs,
		OutputFields:         outputFields,
		selectedExprsReverse: selectedExprsReverse,
	}, nil
}
