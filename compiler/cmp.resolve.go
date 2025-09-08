package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) resolve(node sqlparser.SQLNode, cmpType COMPILER) (string, error) {
	if x, ok := node.(sqlparser.SelectExpr); ok {
		return cmp.selectExpr(x, cmpType)
	}
	if x, ok := node.(*sqlparser.ColName); ok {
		return cmp.colName(x, cmpType)
	}
	if x, ok := node.(*sqlparser.BinaryExpr); ok {
		return cmp.binaryExpr(x, cmpType)
	}
	if x, ok := node.(*sqlparser.FuncExpr); ok {
		return cmp.funcExpr(x, cmpType)
	}
	if x,ok:=node.(*sqlparser.SQLVal);ok {
		return cmp.sqlVal(x, cmpType)
	}
	panic(fmt.Sprintf("Not support %T", node))
}

func (cmp *compiler) binaryExpr(expr *sqlparser.BinaryExpr, cmpType COMPILER) (string, error) {
	strLeft, err := cmp.resolve(expr.Left, C_EXPR)
	if err != nil {
		return "", err
	}
	strRight, err := cmp.resolve(expr.Right, C_EXPR)
	if err != nil {
		return "", err
	}
	return strLeft + expr.Operator + strRight, nil
}
func (cmp *compiler) selectExpr(expr sqlparser.SelectExpr, cmpType COMPILER) (string, error) {
	if x, ok := expr.(*sqlparser.AliasedExpr); ok {
		tableAlias := x.As.String()
		field := ""
		if x.As.IsEmpty() {
			if len(cmp.dict.Tables) > 1 {
				return "", fmt.Errorf("sql command is error")
			}
			tableAlias = cmp.dict.TableAlias[cmp.dict.Tables[0]]

		}
		if c, ok := x.Expr.(*sqlparser.ColName); ok {
			field = c.Name.String()
			matchField := strings.ToLower(fmt.Sprintf("%s.%s", cmp.dict.Tables[0], c.Name.String()))
			if retField, ok := cmp.dict.Field[matchField]; ok {
				if cmpType != C_SELECT {
					return retField, nil
				}
				return retField + " " + cmp.dialect.Quote(cmp.dict.StructField[matchField].Name), nil
			} else {
				if cmpType != C_SELECT {
					return cmp.dialect.Quote(tableAlias, field), nil
				}
				return cmp.dialect.Quote(tableAlias, field) + " " + cmp.dialect.Quote(field), nil
			}

		} else {
			expr, err := cmp.resolve(x.Expr, cmpType)
			if err != nil {
				return "", err
			}
			if x.As.IsEmpty() {
				return expr, nil
			} else {
				if cmpType != C_SELECT {
					return expr, nil
				}
				return expr + " " + cmp.dialect.Quote(x.As.String()), nil
			}

		}

	}
	panic(fmt.Sprintf("Not support %T", expr))
}
