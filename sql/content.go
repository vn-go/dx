package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type contentNode struct {
}

func (c *contentNode) toText(node sqlparser.SQLNode) string {
	switch n := node.(type) {
	case *sqlparser.ColName:
		return c.colNameToText(n)
	case *sqlparser.JoinTableExpr:
		return c.joinTableExprToText(n)
	case *sqlparser.AliasedTableExpr:
		return c.aliasedTableExprToText(n)
	case sqlparser.TableName:
		return c.tableNameToText(n)
	case *sqlparser.Subquery:
		return c.subqueryToText(n)
	case *sqlparser.Select:
		return c.selectToText(n)
	case *sqlparser.AliasedExpr:
		return c.aliasedExprToText(n)
	case sqlparser.TableExprs:
		return c.tableExprs(n)
	case *sqlparser.Where:
		return c.whereToText(n)
	case *sqlparser.ComparisonExpr:
		return c.binaryExprToText(n.Left, n.Right, n.Operator)
	case sqlparser.JoinCondition:
		return c.joinConditionToText(n)
	case *sqlparser.SQLVal:
		if n.Type == sqlparser.ValArg {
			return "?"
		}
		return string(n.Val)
	default:
		panic((fmt.Sprintf("unhandled node type: %T. See contentNode.toText. File %s", n, `sql\content.go`)))
	}

}

func (c *contentNode) joinConditionToText(n sqlparser.JoinCondition) string {
	return c.toText(n.On)

}

func (c *contentNode) binaryExprToText(left sqlparser.Expr, right sqlparser.Expr, operator string) string {
	return fmt.Sprintf("%s %s %s", c.toText(left), operator, c.toText(right))
}

func (c *contentNode) whereToText(n *sqlparser.Where) string {
	return c.toText(n.Expr)
}

func (c *contentNode) tableExprs(n sqlparser.TableExprs) string {
	items := []string{}
	for _, x := range n {
		items = append(items, c.toText(x))
	}
	return strings.Join(items, ", ")
}

func (c *contentNode) aliasedExprToText(n *sqlparser.AliasedExpr) string {
	if n.As.IsEmpty() {
		return c.toText(n.Expr)
	} else {
		return fmt.Sprintf("%s AS %s", c.toText(n.Expr), n.As.String())
	}

}

func (c *contentNode) selectToText(n *sqlparser.Select) string {
	selects := []string{}
	items := []string{}
	for _, expr := range n.SelectExprs {
		selects = append(selects, c.toText(expr))
	}
	items = append(items, fmt.Sprintf("select %s", strings.Join(selects, ", ")))
	if n.From != nil {
		items = append(items, fmt.Sprintf("from %s", c.toText(n.From)))
	}
	if n.Where != nil {
		items = append(items, fmt.Sprintf("where %s", c.toText(n.Where)))
	}
	return strings.Join(items, "\n")

}

func (c *contentNode) subqueryToText(n *sqlparser.Subquery) string {
	if n.Select == nil {
		return ""
	}
	return "(\t\t\n" + c.toText(n.Select) + "\n\t\t)"
}

func (c *contentNode) tableNameToText(n sqlparser.TableName) string {
	if n.Qualifier.IsEmpty() {
		return n.Name.String()
	}
	return fmt.Sprintf("%s.%s", n.Qualifier.String(), n.Name.String())
}

func (c *contentNode) aliasedTableExprToText(n *sqlparser.AliasedTableExpr) string {
	if n.As.IsEmpty() {
		return c.toText(n.Expr)
	}
	return fmt.Sprintf("%s AS %s", c.toText(n.Expr), n.As.String())
}

func (c *contentNode) joinTableExprToText(n *sqlparser.JoinTableExpr) string {
	left := c.toText(n.LeftExpr)
	right := c.toText(n.RightExpr)
	on := c.toText(n.Condition)
	return fmt.Sprintf("%s %s %s ON %s", left, n.Join, right, on)
}

func (c *contentNode) colNameToText(n *sqlparser.ColName) string {
	if n.Qualifier.IsEmpty() {
		return n.Name.String()
	}
	ret := fmt.Sprintf("%s.%s", n.Qualifier.Name.String(), n.Name.String())
	return ret
}

var contents = &contentNode{}
