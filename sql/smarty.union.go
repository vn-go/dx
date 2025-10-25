package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type union struct{}

func (u *union) extractUnionInfo(selectStm *sqlparser.Select, subSetInfoList map[string]subsetInfo) (string, error) {
	for _, x := range selectStm.SelectExprs {
		if fn := detect[*sqlparser.FuncExpr](x); fn != nil {
			if fn.Name.Lowered() == "union" {
				ret, err := u.resolveUnion(fn.Exprs, subSetInfoList)
				if err != nil {
					return "", err
				}
				return ret, nil
			}
		}
	}
	return "", nil
}

func (u *union) resolveUnion(exprs sqlparser.SelectExprs, subSetInfoList map[string]subsetInfo) (string, error) {
	if len(exprs) == 0 {
		return "", newCompilerError(ERR_SYNTAX, "union requires at least two subqueries")
	}
	items := []string{}
	if n := detect[*sqlparser.BinaryExpr](exprs[0]); n != nil {
		if n.Operator != "+" && n.Operator != "*" {
			return "", newCompilerError(ERR_SYNTAX, "union accept only + or * operator")
		}
		l, err := u.getFromExpr(n.Left, subSetInfoList)
		if err != nil {
			return "", err
		}
		items = append(items, l...)
		if n.Operator == "+" { // + is union all
			items = append(items, "UNION ALL")
		} else {
			items = append(items, "UNION")
		}
		r, err := u.getFromExpr(n.Right, subSetInfoList)
		if err != nil {
			return "", err
		}

		items = append(items, r...)

	}
	return strings.Join(items, " "), nil
}

func (u *union) getFromExpr(expr sqlparser.Expr, subSetInfoList map[string]subsetInfo) ([]string, error) {
	switch t := expr.(type) {
	case *sqlparser.ColName:
		subsetsName := t.Name.Lowered()

		if _, ok := subSetInfoList[subsetsName]; !ok {
			return nil, newCompilerError(ERR_SYNTAX, "subsets %s not found", subsetsName)
		}
		return []string{subSetInfoList[subsetsName].query}, nil
	case *sqlparser.BinaryExpr:
		items := []string{}
		if t.Operator != "+" && t.Operator != "*" {
			return nil, newCompilerError(ERR_SYNTAX, "union accept only + or * operator")
		}
		l, err := u.getFromExpr(t.Left, subSetInfoList)
		if err != nil {
			return nil, err
		}
		items = append(items, l...)
		if t.Operator == "+" { // + is union all
			items = append(items, "UNION ALL")
		} else {
			items = append(items, "UNION")
		}
		r, err := u.getFromExpr(t.Right, subSetInfoList)
		if err != nil {
			return nil, err
		}
		return append(items, r...), nil
	default:
		panic(fmt.Sprintf("unimplemented: %T. see union.getFromExpr", t))
	}

}

var unions = &union{}
