package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// select.selects.go
func (s selectors) selects(expr *sqlparser.Select, injector *injector) (*compilerResult, error) {
	ret := compilerResult{}
	sql := sqlComplied{}

	r, err := froms.resolve(expr.From, injector)
	if err != nil {
		return nil, err
	}
	ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
	sql.source = r.Content
	itemSelectors := []string{}
	for _, x := range expr.SelectExprs {
		r, err = s.selectExpr(x, injector)
		if err != nil {
			return nil, err
		}
		itemSelectors = append(itemSelectors, r.Content+" "+injector.dialect.Quote(r.AliasOfContent))
		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		ret.selectedExprs = internal.UnionMap(ret.selectedExprs, r.selectedExprs)
		ret.selectedExprsReverse = internal.UnionMap(ret.selectedExprsReverse, r.selectedExprsReverse)
	}
	sql.selector = strings.Join(itemSelectors, ", ")
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
				havingItems = append(havingItems, *&r.Content)
			} else {
				resultOfWhere = append(resultOfWhere, *&r.Content)
			}
		}
		if len(resultOfWhere) > 0 {
			sql.filter = strings.Join(resultOfWhere, " AND ")
		}
		if len(havingItems) > 0 {
			sql.having = strings.Join(havingItems, " AND ")
			goupByItems := []string{}
			checkGroupBy := map[string]bool{}
			for k, v := range ret.selectedExprsReverse {
				if k == "" { // not not hav alias skip it
					continue
				}
				if !v.IsInAggregateFunc {
					if _, ok := checkGroupBy[v.Expr]; !ok {
						goupByItems = append(goupByItems, v.Expr)
						checkGroupBy[v.Expr] = true
					}
				}
			}
			if len(goupByItems) > 0 {
				sql.groupBy = strings.Join(goupByItems, ", ")
			}
		}
	}

	ret.Content = sql.String()
	ret.Args = injector.args
	return &ret, nil
}

func (s selectors) selectExpr(expr sqlparser.SelectExpr, injector *injector) (*compilerResult, error) {
	switch x := expr.(type) {
	case *sqlparser.StarExpr:
		return s.starExpr(x, injector)
	case *sqlparser.AliasedExpr:
		r, err := exp.resolve(x.Expr, injector, CMP_SELECT)
		if err != nil {
			return nil, err
		}
		if x.As.IsEmpty() && r.IsExpression {
			return nil, newCompilerError("Please add a name (alias) for the expression '%s'.", r.OriginalContent)
		} else if !x.As.IsEmpty() {
			r.AliasOfContent = x.As.String()
		}
		r.selectedExprsReverse.merge(dictionaryFields{
			x.As.Lowered(): &dictionaryField{
				Expr:              r.Content,
				IsInAggregateFunc: r.IsInAggregateFunc,
				Alias:             x.As.String(),
			},
		})

		return r, nil
	default:
		panic(fmt.Sprintf("unimplemented: %T. See selectors.selectExpr, file %s", x, `sql\select.selects.go`))
	}

}

func (s selectors) starExpr(expr *sqlparser.StarExpr, injector *injector) (*compilerResult, error) {
	strSelectItems := []string{}
	if expr.TableName.IsEmpty() {
		for _, x := range injector.dict.entities {
			for _, col := range x.Cols {
				aliasField := injector.dialect.Quote(col.Field.Name)
				if len(injector.dict.entities) > 1 { // if there are more than one entity, we need to add entity name to alias
					aliasField = injector.dialect.Quote(x.EntityType.Name() + "_" + col.Field.Name)
				}
				strSelectItems = append(strSelectItems, injector.dialect.Quote(x.TableName, col.Name)+" "+aliasField)
				refFieldKey := strings.ToLower(fmt.Sprintf("%s.%s", x.EntityType.Name(), col.Field.Name))
				if _, ok := injector.fields[refFieldKey]; !ok {
					injector.fields[refFieldKey] = refFieldInfo{
						EntityName:      x.EntityType.Name(),
						EntityFieldName: col.Field.Name,
					}
				}
			}
		}
	} else {
		if ent, ok := injector.dict.entities[strings.ToLower(expr.TableName.Name.String())]; ok {
			for _, col := range ent.Cols {
				strSelectItems = append(strSelectItems, injector.dialect.Quote(ent.TableName, col.Name)+" "+injector.dialect.Quote(col.Field.Name))
				refFieldKey := strings.ToLower(fmt.Sprintf("%s.%s", ent.EntityType.Name(), col.Field.Name))
				if _, ok := injector.fields[refFieldKey]; !ok {
					injector.fields[refFieldKey] = refFieldInfo{
						EntityName:      ent.EntityType.Name(),
						EntityFieldName: col.Field.Name,
					}
				}
			}
		} else {
			return nil, newCompilerError("datasource %s not found", expr.TableName.Name.String())
		}
	}
	strSelect := strings.Join(strSelectItems, ", ")
	return &compilerResult{
		Content: strSelect,
		Args:    nil,
		Fields:  injector.fields,
	}, nil
}
