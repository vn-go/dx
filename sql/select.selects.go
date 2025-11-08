package sql

import (
	"fmt"

	sortTexts "sort"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func sortStrings(items []string) []string {
	sorted := make([]string, len(items))
	copy(sorted, items)
	sortTexts.Strings(sorted)
	return sorted
}

type getSelectStatementResult struct {
	selectStatement types.SelectStatement
	compilerResult  compilerResult
	args            argsBoard
}

func (s selectors) getSelectStatement(expr *sqlparser.Select, injector *injector, cmpType CMP_TYP) (*getSelectStatementResult, error) {
	ret := compilerResult{
		OutputFields: []outputField{},
	}
	selectStatement := types.SelectStatement{}
	argsB := argsBoard{
		Source:   arguments{},
		Selector: arguments{},
		Filter:   arguments{},
		Sort:     arguments{},
		Having:   arguments{},
		GroupBy:  arguments{},
	}

	r, err := froms.resolve(expr.From, injector)
	if err != nil {
		return nil, err
	}
	ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
	selectStatement.Source = r.Content
	itemSelectors := []string{}
	selectedExprsReverse := &dictionaryFields{}
	for _, x := range expr.SelectExprs {

		r, err = s.selectExpr(x, injector, cmpType, selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		if _, ok := x.(*sqlparser.StarExpr); ok {
			itemSelectors = append(itemSelectors, r.Content)
		} else {
			if cmpType == CMP_SELECT {
				itemSelectors = append(itemSelectors, r.Content+" "+injector.dialect.Quote(r.AliasOfContent))
			} else {
				if r.AliasOfContent != "" {
					itemSelectors = append(itemSelectors, r.Content+" "+injector.dialect.Quote(r.AliasOfContent))
				} else {
					itemSelectors = append(itemSelectors, r.Content)
				}

			}

		}

		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		ret.selectedExprs = internal.UnionMap(ret.selectedExprs, r.selectedExprs)
		ret.selectedExprsReverse = internal.UnionMap(ret.selectedExprsReverse, r.selectedExprsReverse)
		ret.IsInAggregateFunc = ret.IsInAggregateFunc || r.IsInAggregateFunc
		ret.OutputFields = append(ret.OutputFields, r.OutputFields...)
		argsB.Selector = append(argsB.Selector, r.Args...)
	}
	selectStatement.Selector = strings.Join(itemSelectors, ", ")

	//goupByItems := []string{}
	//checkGroupBy := map[string]bool{}
	havingItems := []string{}
	groupKeys := []string{}
	groupMap := map[string]string{}
	if expr.Where != nil {
		resultOfWhere := []string{}

		nodes := where.splitAndExpr(expr.Where.Expr)
		for _, node := range nodes {
			//field Expr sqlparser.Expr
			// whereselectedExprsReverse := &dictionaryFields{}
			r, err = where.resolve(node.(sqlparser.Expr), injector, selectedExprsReverse)
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
				selectedExprsReverse.merge(r.selectedExprsReverse) // "Fields which do not belong to an aggregate function must be added to the GROUP BY clause."

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

			for k, v := range *selectedExprsReverse {
				if k == "" || v.IsInAggregateFunc { // not not hav alias skip it
					continue
				}
				if _, ok := groupMap[v.Expr]; !ok {
					groupKeys = append(groupKeys, v.Expr)
					groupMap[v.Expr] = v.Expr
				}

			}

		}
	}

	// detect if is need to add group by
	if len(havingItems) > 0 || ret.IsInAggregateFunc {
		for k, v := range *selectedExprsReverse {
			if k == "" || v.IsInAggregateFunc { // not not hav alias skip it
				continue
			}
			if _, ok := groupMap[v.Expr]; !ok {
				groupKeys = append(groupKeys, v.Expr)

				groupMap[v.Expr] = v.Expr
			}
		}
	}
	if expr.GroupBy != nil {

		r, err := groups.resolve(expr.GroupBy, injector, selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		groupKeys = append(groupKeys, r.Content)
		groupMap[r.Content] = r.Content
		//goupByItems = append(goupByItems, r.Content)

	}

	if len(groupKeys) > 0 {
		goupByItems := []string{}
		groupKeys = sortStrings(groupKeys)
		for _, x := range groupKeys {
			if y, ok := groupMap[x]; ok {
				goupByItems = append(goupByItems, y)

			}
		}
		selectStatement.GroupBy = strings.Join(goupByItems, ", ")
	}
	if expr.OrderBy != nil {
		r, err := sort.resolveOrderBy(expr.OrderBy, injector, &ret.selectedExprsReverse)

		if err != nil {
			return nil, err
		}
		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		selectStatement.Sort = r.Content
	}
	if expr.Limit != nil {
		//tmpInjector := newInjector(injector.dialect, make([]string, 0))

		if expr.Limit.Offset != nil {
			offset, err := exp.resolve(expr.Limit.Offset, injector, CMP_SELECT, &ret.selectedExprsReverse)
			if err != nil {
				return nil, err
			}

			selectStatement.Offset = &types.SelectStatementArg{
				Content: offset.Content,
			}

		}
		if expr.Limit.Rowcount != nil {
			limit, err := exp.resolve(expr.Limit.Rowcount, injector, CMP_SELECT, &ret.selectedExprsReverse)
			if err != nil {
				return nil, err
			}

			selectStatement.Limit = &types.SelectStatementArg{
				Content: limit.Content,
			}
			// ret.limit = smartier.ToText(expr.Limit.Rowcount)
		}

	}
	ret.Content = injector.dialect.GetSelectStatement(selectStatement)
	ret.Args = injector.args

	return &getSelectStatementResult{
		selectStatement: selectStatement,
		compilerResult:  ret,
		args:            argsB,
	}, nil
}

// select.selects.go
func (s selectors) selects(expr *sqlparser.Select, injector *injector, cmpType CMP_TYP) (*compilerResult, error) {
	r, err := s.getSelectStatement(expr, injector, cmpType)
	if err != nil {
		return nil, err
	}
	return &r.compilerResult, nil
}

func (s selectors) selectExpr(expr sqlparser.SelectExpr, injector *injector, cmpType CMP_TYP, selectedExprsReverse *dictionaryFields) (*compilerResult, error) {
	switch x := expr.(type) {
	case *sqlparser.StarExpr:
		return s.starExpr(x, injector)
	case *sqlparser.AliasedExpr:
		x.Expr = exp.detectIsConcatFunc(x.Expr, injector.dict.fields)
		r, err := exp.resolve(x.Expr, injector, CMP_SELECT, selectedExprsReverse)
		if err != nil {
			return nil, err
		}
		if x.As.IsEmpty() && r.IsExpression && cmpType == CMP_SELECT {
			return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "Please add a name (alias) for the expression '%s'.", r.OriginalContent)
		} else if !x.As.IsEmpty() && cmpType == CMP_SELECT {
			r.AliasOfContent = x.As.String()
		}

		if x.As.IsEmpty() {
			if cmpType == CMP_SUBQUERY {
				// if r.AliasOfContent == "" {

				// }
				return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "Please add a name (alias) for the expression '%s'.", r.OriginalContent)
			}

			// if r.IsInSubquery {
			// 	// if r.AliasOfContent == "" {

			// 	// }
			// 	return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "Please add a name (alias) for the expression '%s'.", r.OriginalContent)
			// }
			selectedExprsReverse.merge(r.selectedExprsReverse)
		} else {
			if cmpType == CMP_SELECT || cmpType == CMP_SUBQUERY {
				(*selectedExprsReverse)[x.As.Lowered()] = &dictionaryField{
					Expr:              r.Content,
					IsInAggregateFunc: r.IsInAggregateFunc,
					Alias:             x.As.String(),
				}
			}
		}
		aliasField := x.As.String()
		if x.As.IsEmpty() {
			aliasField = r.AliasOfContent
		}
		if aliasField == "" {
			return nil, newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "Please add a name (alias) for the expression '%s'.", r.OriginalContent)
		}
		r.selectedExprs.merge(dictionaryFields{
			strings.ToLower(aliasField): &dictionaryField{
				Expr:              r.Content,
				IsInAggregateFunc: r.IsInAggregateFunc,
				Alias:             aliasField,
			},
		})
		ret := &compilerResult{
			OriginalContent:      r.OriginalContent,
			Content:              r.Content,
			AliasOfContent:       aliasField,
			selectedExprs:        r.selectedExprs,
			nonAggregateFields:   r.nonAggregateFields,
			selectedExprsReverse: *selectedExprsReverse,
			IsInAggregateFunc:    r.IsInAggregateFunc,
			Fields:               r.Fields,
			OutputFields:         []outputField{},
			Args:                 r.Args,
		}

		if !x.As.IsEmpty() {
			ret.OutputFields = append(ret.OutputFields, outputField{
				Name:         x.As.String(),
				IsCalculated: r.IsExpression,
				FieldType:    r.ResultType,
				Expression:   r.Content,
			})
		} else {
			ret.OutputFields = append(ret.OutputFields, outputField{
				Name:         aliasField,
				IsCalculated: r.IsExpression,
				FieldType:    r.ResultType,
				Expression:   r.Content,
			})

		}
		if cmpType == CMP_SELECT {
			(*selectedExprsReverse)[strings.ToLower(aliasField)] = &dictionaryField{
				Expr:              r.Content,
				IsInAggregateFunc: r.IsInAggregateFunc,
				Alias:             aliasField,
			}
		}
		if len(ret.selectedExprs) == 0 {
			panic("selectedExprs is empty")
		}
		return ret, nil

	default:
		panic(fmt.Sprintf("unimplemented: %T. See selectors.selectExpr, file %s", x, `sql\select.selects.go`))
	}

}
