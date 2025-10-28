package sql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type crossTab struct {
}

func (f *crossTab) resolve(alias *sqlparser.AliasedExpr, fn *sqlparser.FuncExpr, fieldAliasMap map[string]string) (string, error) {
	// if alias.As.IsEmpty() {
	// 	return "", newCompilerError(ERR_EXPRESION_REQUIRE_ALIAS, "expression '%s' require alias", smartier.ToText(fn))
	// }
	// determine range:
	var fnFor *sqlparser.FuncExpr
	var fnSelect *sqlparser.FuncExpr

	for _, x := range fn.Exprs {
		if fx := detect[*sqlparser.FuncExpr](x); fx != nil {
			if fx.Name.Lowered() == "for" {
				fnFor = fx

			}
		}
		if fx := detect[*sqlparser.FuncExpr](x); fx != nil {
			if fx.Name.Lowered() == "select" {
				fnSelect = fx

			}
		}
	}
	if fnFor == nil || fnSelect == nil {
		return "", newCompilerError(ERR_SYNTAX, smartier.ToText(fn))
	}
	return f.resolveFull(fnFor, fnSelect, fieldAliasMap)
}

func (f *crossTab) resolveFull(fnFor, fnSelect *sqlparser.FuncExpr, fieldAliasMap map[string]string) (string, error) {
	iterator := fnFor.Exprs[0]
	iteratorExpr, ok := iterator.(*sqlparser.AliasedExpr)
	if !ok {
		return "", newCompilerError(ERR_FIELD_REQUIRE_ALIAS, "Expression '%s' require alias", smartier.ToText(iterator))
	}

	minValNode, ok := fnFor.Exprs[1].(*sqlparser.AliasedExpr).Expr.(*sqlparser.SQLVal)

	if !ok {
		return "", newCompilerError(ERR_SYNTAX, "Invalid min value for 'for' function. Expr '%s' in '%s'", smartier.ToText(fnFor.Exprs[1]), smartier.ToText(fnFor))
	}
	if minValNode.Type != sqlparser.IntVal {
		return "", newCompilerError(ERR_SYNTAX, "Invalid min value for 'for' function. Expr '%s' in '%s'", smartier.ToText(fnFor.Exprs[1]), smartier.ToText(fnFor))
	}
	minVal, err := internal.Helper.ToIntFormBytes(minValNode.Val)
	if err != nil {
		return "", newCompilerError(ERR_SYNTAX, "Invalid min value for 'for' function. Expr '%s' in '%s'", smartier.ToText(fnFor.Exprs[1]), smartier.ToText(fnFor))
	}
	maxValNode, ok := fnFor.Exprs[2].(*sqlparser.AliasedExpr).Expr.(*sqlparser.SQLVal)
	if !ok {
		return "", newCompilerError(ERR_SYNTAX, "Invalid max value for 'for' function. Expr '%s' in '%s'", smartier.ToText(fnFor.Exprs[2]), smartier.ToText(fnFor))
	}
	if maxValNode.Type != sqlparser.IntVal {
		return "", newCompilerError(ERR_SYNTAX, "Invalid max value for 'for' function. Expr '%s' in '%s'", smartier.ToText(fnFor.Exprs[2]), smartier.ToText(fnFor))
	}
	maxVal, err := internal.Helper.ToIntFormBytes(maxValNode.Val)
	if err != nil {
		return "", newCompilerError(ERR_SYNTAX, "Invalid max value for 'for' function. Expr '%s' in '%s'", smartier.ToText(fnFor.Exprs[2]), smartier.ToText(fnFor))
	}
	if minVal > maxVal {
		return "", newCompilerError(ERR_SYNTAX, "Invalid range for 'for' function. Min value '%d' is greater than max value '%d' in '%s'", minVal, maxVal, smartier.ToText(fnFor))
	}
	return f.build(iteratorExpr, fnSelect.Exprs, minVal, maxVal, fieldAliasMap)
}

func (f *crossTab) build(iteratorExpr *sqlparser.AliasedExpr, selectExpr sqlparser.SelectExprs, minVal int, maxVal int, fieldAliasMap map[string]string) (string, error) {
	items := []string{}
	alias := iteratorExpr.As.String()
	for _, x := range selectExpr {
		masterAlais := x.(*sqlparser.AliasedExpr).As.String()
		if fx, ok := x.(*sqlparser.AliasedExpr).Expr.(*sqlparser.FuncExpr); ok {
			for i := minVal; i <= maxVal; i++ {
				node := &sqlparser.BinaryExpr{
					Operator: "=",
					Left:     iteratorExpr.Expr,
					Right:    &sqlparser.SQLVal{Type: sqlparser.IntVal, Val: []byte(strconv.Itoa(i))},
				}

				ifFn := &sqlparser.FuncExpr{
					Name: sqlparser.NewColIdent("if"),
					Exprs: []sqlparser.SelectExpr{
						&sqlparser.AliasedExpr{
							Expr: node,
						},
						selectExpr[0].(*sqlparser.AliasedExpr).Expr.(*sqlparser.FuncExpr).Exprs[0],
					},
				}
				aggCrossTab := &sqlparser.FuncExpr{
					Name: fx.Name,
					Exprs: []sqlparser.SelectExpr{
						&sqlparser.AliasedExpr{
							Expr: ifFn,
						},
					},
				}
				aliasExpr := &sqlparser.AliasedExpr{
					Expr: aggCrossTab,
					As:   sqlparser.NewColIdent(fmt.Sprintf("%s%s%d", masterAlais, alias, i)),
				}
				// fmt.Printf("node: %s\n", smartier.ToText(aliasExpr))
				items = append(items, smartier.ToText(aliasExpr))
			}
		} else {
			return "", newCompilerError(ERR_SYNTAX, "Invalid expression '%s' in 'for' function", smartier.ToText(x))
		}
	}

	return strings.Join(items, ","), nil
}

var crossTabs = &crossTab{}
