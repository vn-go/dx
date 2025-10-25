package sql

import (
	"fmt"

	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// select.selects.go
func (s selectors) selects(expr *sqlparser.Select, injector *injector) (*compilerResult, error) {
	ret := compilerResult{}
	selectStatement := types.SelectStatement{}

	r, err := froms.resolve(expr.From, injector)
	if err != nil {
		return nil, err
	}
	ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
	selectStatement.Source = r.Content
	itemSelectors := []string{}
	for _, x := range expr.SelectExprs {

		r, err = s.selectExpr(x, injector)
		if err != nil {
			return nil, err
		}
		if _, ok := x.(*sqlparser.StarExpr); ok {
			itemSelectors = append(itemSelectors, r.Content)
		} else {
			itemSelectors = append(itemSelectors, r.Content+" "+injector.dialect.Quote(r.AliasOfContent))
		}

		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		ret.selectedExprs = internal.UnionMap(ret.selectedExprs, r.selectedExprs)
		ret.selectedExprsReverse = internal.UnionMap(ret.selectedExprsReverse, r.selectedExprsReverse)
	}
	selectStatement.Selector = strings.Join(itemSelectors, ", ")
	goupByItems := []string{}
	checkGroupBy := map[string]bool{}
	if expr.Where != nil {
		resultOfWhere := []string{}
		havingItems := []string{}
		nodes := where.splitAndExpr(expr.Where.Expr)
		for _, node := range nodes {
			//field Expr sqlparser.Expr

			r, err = where.resolve(node.(sqlparser.Expr), injector, ret.selectedExprsReverse)
			if err != nil {
				return nil, err
			}

			if r.IsInAggregateFunc {
				/*
					If after compiling the expression, it is an aggregate function,
					it means it cannot be used in the WHERE clause.
					So we need to add it to the HAVING clause
				*/

				havingItems = append(havingItems, r.Content)
				ret.selectedExprsReverse.merge(r.selectedExprsReverse) // "Fields which do not belong to an aggregate function must be added to the GROUP BY clause."

			} else {
				resultOfWhere = append(resultOfWhere, r.Content)
			}
			ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		}
		if len(resultOfWhere) > 0 {
			selectStatement.Filter = strings.Join(resultOfWhere, " AND ")
		}
		if len(havingItems) > 0 {
			selectStatement.Having = strings.Join(havingItems, " AND ")

			for k, v := range ret.selectedExprsReverse {
				if k == "" { // not not hav alias skip it
					continue
				}
				if !v.IsInAggregateFunc {
					if v.Children != nil && len(*v.Children) > 0 {
						for k1, child := range *v.Children {
							if k1 == "" { // not not hav alias skip it
								continue
							}
							if _, ok := checkGroupBy[child.Expr]; !ok {

								goupByItems = append(goupByItems, child.Expr)
								checkGroupBy[child.Expr] = true
							}
						}

					} else {
						if _, ok := checkGroupBy[v.Expr]; !ok {

							goupByItems = append(goupByItems, v.Expr)
							checkGroupBy[v.Expr] = true
						}
					}
				}
			}

		}
	}
	// detect if is need to add group by
	for k, v := range ret.selectedExprsReverse {
		if k == "" { // not not hav alias skip it
			continue
		}
		if !v.IsInAggregateFunc {
			if v.Children != nil && len(*v.Children) > 0 {
				for k1, child := range *v.Children {
					if k1 == "" { // not not hav alias skip it
						continue
					}
					if _, ok := checkGroupBy[child.Expr]; !ok {

						goupByItems = append(goupByItems, child.Expr)
						checkGroupBy[child.Expr] = true
					}
				}

			} else {
				if _, ok := checkGroupBy[v.Expr]; !ok {

					goupByItems = append(goupByItems, v.Expr)
					checkGroupBy[v.Expr] = true
				}
			}
		}
	}
	if len(goupByItems) > 0 {
		selectStatement.GroupBy = strings.Join(goupByItems, ", ")
	}
	if expr.OrderBy != nil {
		r, err := sort.resolveOrderBy(expr.OrderBy, injector, ret.selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		selectStatement.Sort = r.Content
	}
	if expr.GroupBy != nil {

		r, err := groups.resolve(expr.GroupBy, injector, ret.selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		selectStatement.GroupBy = r.Content
	}
	ret.Content = injector.dialect.GetSelectStatement(selectStatement)
	ret.Args = injector.args
	return &ret, nil
}

func (s selectors) selectExpr(expr sqlparser.SelectExpr, injector *injector) (*compilerResult, error) {
	switch x := expr.(type) {
	case *sqlparser.StarExpr:
		return s.starExpr(x, injector)
	case *sqlparser.AliasedExpr:
		r, err := exp.resolve(x.Expr, injector, CMP_SELECT, dictionaryFields{})
		if err != nil {
			return nil, err
		}
		if x.As.IsEmpty() && r.IsExpression {
			return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "Please add a name (alias) for the expression '%s'.", r.OriginalContent)
		} else if !x.As.IsEmpty() {
			r.AliasOfContent = x.As.String()
		}
		return &compilerResult{
			OriginalContent:    r.OriginalContent,
			Content:            r.Content,
			AliasOfContent:     r.AliasOfContent,
			selectedExprs:      r.selectedExprs,
			nonAggregateFields: r.nonAggregateFields,
			selectedExprsReverse: dictionaryFields{
				x.As.Lowered(): &dictionaryField{
					Expr:              r.Content,
					IsInAggregateFunc: r.IsInAggregateFunc,
					Alias:             x.As.String(),
					Children:          &r.selectedExprsReverse,
				},
			},
			IsInAggregateFunc: r.IsInAggregateFunc,
			Fields:            r.Fields,
		}, nil
		// r.selectedExprsReverse.merge(dictionaryFields{
		// 	x.As.Lowered(): &dictionaryField{
		// 		Expr:              r.Content,
		// 		IsInAggregateFunc: r.IsInAggregateFunc,
		// 		Alias:             x.As.String(),
		// 		Children:          &r.selectedExprsReverse,
		// 	},
		// })

		//return r, nil

	default:
		panic(fmt.Sprintf("unimplemented: %T. See selectors.selectExpr, file %s", x, `sql\select.selects.go`))
	}

}

func (s selectors) starExpr(expr *sqlparser.StarExpr, injector *injector) (*compilerResult, error) {
	strSelectItems := []string{}
	selectedExprs := dictionaryFields{}
	if expr.TableName.IsEmpty() {
		i := 1
		for _, x := range injector.dict.entities {
			aliasTable := fmt.Sprintf("T%d", i)
			for _, col := range x.Cols {

				aliasField := injector.dialect.Quote(col.Field.Name)
				if len(injector.dict.entities) > 1 { // if there are more than one entity, we need to add entity name to alias
					aliasField = injector.dialect.Quote(x.EntityType.Name() + "_" + col.Field.Name)
				}
				strSelectItems = append(strSelectItems, injector.dialect.Quote(x.TableName, col.Name)+" "+aliasField)
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
			}
		} else {
			return nil, newCompilerError(ERR_DATASET_NOT_FOUND, "dataset %s not found", expr.TableName.Name.String())
		}
	}
	strSelect := strings.Join(strSelectItems, ", ")
	return &compilerResult{
		Content:       strSelect,
		Args:          nil,
		Fields:        injector.fields,
		selectedExprs: selectedExprs,
	}, nil
}
