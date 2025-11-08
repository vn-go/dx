package sql

import (
	"fmt"

	"github.com/vn-go/dx/sqlparser"
)

type tableApply struct {
}

func (t *tableApply) resolve(node sqlparser.SQLNode, table string) sqlparser.SQLNode {
	switch n := node.(type) {
	case *sqlparser.AliasedExpr:
		return t.aliasedExpr(n, table)
	case *sqlparser.FuncExpr:
		return t.funcExpr(n, table)
	case *sqlparser.StarExpr:
		return &sqlparser.StarExpr{
			TableName: sqlparser.TableName{
				Name: sqlparser.NewTableIdent(table),
			},
		}
	default:

		panic(fmt.Sprintf("not implement %T. ref tableApply.resolve file%s ", n, `sql\smarty.smarty.tableApplier.go`))
	}
}

func (t *tableApply) funcExpr(n *sqlparser.FuncExpr, table string) sqlparser.SQLNode {
	panic("unimplemented")
}

func (t *tableApply) aliasedExpr(n *sqlparser.AliasedExpr, table string) sqlparser.SQLNode {
	n.Expr = t.expr(n.Expr, table).(sqlparser.Expr)
	return n
}

func (t *tableApply) expr(expr sqlparser.SQLNode, table string) sqlparser.SQLNode {
	switch n := expr.(type) {
	case *sqlparser.FuncExpr:
		exprs := sqlparser.SelectExprs{}
		for i := 0; i < len(n.Exprs); i++ {

			fx := t.expr(n.Exprs[i], table).(sqlparser.SelectExpr)
			exprs = append(exprs, fx)
		}
		n.Exprs = exprs
		return n
	case *sqlparser.AliasedExpr:
		n.Expr = t.expr(n.Expr, table).(sqlparser.Expr)
		return n
	case *sqlparser.ColName:
		n.Qualifier = t.expr(n.Qualifier, table).(sqlparser.TableName)
		return n
	case sqlparser.TableName:
		n.Name = t.expr(n.Name, table).(sqlparser.TableIdent)
		return n
	case sqlparser.TableIdent:
		if n.IsEmpty() {
			return sqlparser.NewTableIdent(table)
		} else {
			return n
		}

	case *sqlparser.BinaryExpr:
		n.Left = t.expr(n.Left, table).(sqlparser.Expr)
		n.Right = t.expr(n.Right, table).(sqlparser.Expr)
		return n
	case *sqlparser.AndExpr:
		n.Left = t.expr(n.Left, table).(sqlparser.Expr)
		n.Right = t.expr(n.Right, table).(sqlparser.Expr)
		return n
	case *sqlparser.OrExpr:
		n.Left = t.expr(n.Left, table).(sqlparser.Expr)
		n.Right = t.expr(n.Right, table).(sqlparser.Expr)
		return n
	case *sqlparser.NotExpr:
		n.Expr = t.expr(n.Expr, table).(sqlparser.Expr)
		return n
	case *sqlparser.SQLVal:
		return n
	case *sqlparser.StarExpr:
		n.TableName = t.expr(n.TableName, table).(sqlparser.TableName)
		return n
	case *sqlparser.ComparisonExpr:
		n.Left = t.expr(n.Left, table).(sqlparser.Expr)
		n.Right = t.expr(n.Right, table).(sqlparser.Expr)
		return n
	default:
		panic(fmt.Sprintf("not implement %T. ref tableApply.expr file%s ", n, `sql\smarty.smarty.tableApplier.go`))
	}

}

var tableApplier = &tableApply{}

// smarty.AddTableTableName.go
