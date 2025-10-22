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
		itemSelectors = append(itemSelectors, r.Content)
		ret.Fields = internal.UnionMap(ret.Fields, r.Fields)
		ret.selectedExprs = internal.UnionMap(ret.selectedExprs, r.selectedExprs)
		ret.selectedExprsReverse = internal.UnionMap(ret.selectedExprsReverse, r.selectedExprsReverse)
	}
	sql.selector = strings.Join(itemSelectors, ", ")
	if expr.Where != nil {
		r, err = where.resolve(expr.Where.Expr, injector)
		if err != nil {
			return nil, err
		}
		sql.filter = r.Content
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
		return s.aliasedExpr(x, injector)
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
