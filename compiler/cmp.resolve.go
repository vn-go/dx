package compiler

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

func (cmp *compiler) resolve(node sqlparser.SQLNode, cmpType COMPILER, args *internal.SqlArgs) (string, error) {
	if x, ok := node.(sqlparser.SelectExpr); ok {
		return cmp.selectExpr(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.ColName); ok {
		return cmp.colName(x, cmpType)
	}
	if x, ok := node.(*sqlparser.BinaryExpr); ok {
		return cmp.binaryExpr(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.FuncExpr); ok {
		return cmp.funcExpr(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.SQLVal); ok {
		return cmp.sqlVal(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.AliasedTableExpr); ok {
		return cmp.aliasedTableExpr(x, cmpType, args)
	}
	if x, ok := node.(sqlparser.TableName); ok {
		return cmp.tableName(x, cmpType)
	}
	if x, ok := node.(*sqlparser.ComparisonExpr); ok {
		return cmp.comparisonExpr(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.JoinTableExpr); ok {
		return cmp.joinTableExpr(x, cmpType, args)
	}
	if x, ok := node.(sqlparser.JoinCondition); ok {
		return cmp.joinCondition(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.AndExpr); ok {
		return cmp.andExpr(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.OrExpr); ok {
		return cmp.orExpr(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.NotExpr); ok {
		return cmp.notExpr(x, cmpType, args)
	}
	if x, ok := node.(*sqlparser.IsExpr); ok {
		ret, err := cmp.resolve(x.Expr, cmpType, args)
		if err != nil {
			return "", err
		}
		return ret + strings.ToUpper(x.Operator), nil
	}
	// if x, ok := node.(*sqlparser.Subquery); ok {
	// 	ret, err := cmp.resolve(x.Select, C_SELECT, args)
	// 	return ret, err
	// }
	if x, ok := node.(*sqlparser.Union); ok {
		info, err := cmp.getSqlInfoFromUnion(x)
		if err != nil {
			return "", err
		}
		sqlRet, err := cmp.dialect.BuildSql(info)
		if err != nil {
			return "", err
		}

		return sqlRet.Sql, nil
		// leftCompiler, err := newCompilerFromSqlNode(x.Left, cmp.dialect)
		// if err != nil {
		// 	return "", err
		// }
		// leftSql, err := leftCompiler.resolve(x.Left, C_SELECT, args)
		// cmp.args = internal.UnionCompilerArgs(cmp.args, leftCompiler.args)
		// if err != nil {
		// 	return "", err
		// }
		// rightCompiler, err := newCompilerFromSqlNode(x.Right, cmp.dialect)
		// if err != nil {
		// 	return "", err
		// }
		// rightSql, err := rightCompiler.resolve(x.Left, C_SELECT, args)
		// if err != nil {
		// 	return "", err
		// }
		// cmp.args = internal.UnionCompilerArgs(cmp.args, rightCompiler.args)
		// return leftSql + " " + x.Type + " " + rightSql, nil
	}
	if x, ok := node.(*sqlparser.Select); ok {
		info, err := cmp.getSqlInfoBySelect(x)
		if err != nil {
			return "", err
		}
		ret, err := cmp.dialect.BuildSql(info)
		if err != nil {
			return "", err
		}
		return ret.Sql, nil
	}
	if x, ok := node.(*sqlparser.UpdateExpr); ok {
		strExpr, err := cmp.resolve(x.Expr, C_UPDATE, args)
		if err != nil {
			return "", err
		}
		if len(cmp.dict.Tables) == 1 {
			// alias := cmp.dict.TableAlias[cmp.dict.Tables[0]]
			if field, ok := cmp.dict.Field[strings.ToLower(fmt.Sprintf("%s.%s", cmp.dict.Tables[0], x.Name.Name.String()))]; ok {
				return strings.Split(field, ".")[1] + "=" + strExpr, nil
			}
			return cmp.dialect.Quote(x.Name.Name.String()) + "=" + strExpr, nil

		} else {
			panic(fmt.Sprintf("Not support mmulti source update statement, %s", `compiler\cmp.resolve.go`))
		}

		//return "", nil
	}
	if x, ok := node.(*sqlparser.Where); ok {
		return cmp.resolve(x.Expr, cmpType, args)
	}
	if x, ok := node.(*sqlparser.ParenExpr); ok {
		ret, err := cmp.resolve(x.Expr, cmpType, args)
		if err != nil {
			return "", err
		}
		return "( " + ret + " )", nil
	}
	panic(fmt.Sprintf("Not support %T, %s", node, `compiler\cmp.resolve.go`))
}

func (cmp *compiler) binaryExpr(expr *sqlparser.BinaryExpr, cmpType COMPILER, args *internal.SqlArgs) (string, error) {
	strLeft, err := cmp.resolve(expr.Left, C_EXPR, args)
	if err != nil {
		return "", err
	}
	strRight, err := cmp.resolve(expr.Right, C_EXPR, args)
	if err != nil {
		return "", err
	}
	return strLeft + expr.Operator + strRight, nil
}
func (cmp *compiler) selectExpr(expr sqlparser.SelectExpr, cmpType COMPILER, args *internal.SqlArgs) (string, error) {
	if cmp.dict.ExprAlias == nil {
		cmp.dict.ExprAlias = make(map[string]types.OutputExpr)
	}
	if x, ok := expr.(*sqlparser.AliasedExpr); ok {
		resStr, err := cmp.resolve(x.Expr, cmpType, args)
		if _, ok := x.Expr.(*sqlparser.SQLVal); ok {
			if x.As.IsEmpty() {
				return resStr, nil
			} else {
				cmp.dict.ExprAlias[strings.ToLower(x.As.String())] = types.OutputExpr{
					SqlNode: x.Expr,
					Expr: types.FiedlExpression{
						ExprContent: resStr,
						ExprType:    types.EXPR_TYPE(types.EXPR_TYPE_VAL),
					},
				} // resStr
				return resStr + " " + cmp.dialect.Quote(x.As.String()), nil
			}

		}
		if fnNode, ok := x.Expr.(*sqlparser.FuncExpr); ok {
			if fnNode.Name.String() == internal.FnMarkSpecialTextArgs {
				if x.As.IsEmpty() {
					return resStr, nil
				} else {
					return resStr + " " + cmp.dialect.Quote(x.As.String()), nil
				}
			}
		}

		if err != nil {
			return "", err
		}
		if resStr != "" {
			if !x.As.IsEmpty() && cmpType == C_SELECT {
				cmp.dict.ExprAlias[strings.ToLower(x.As.String())] = types.OutputExpr{
					SqlNode: x,
					Expr: types.FiedlExpression{
						ExprContent: resStr,
					},
				} //resStr

				return resStr + " " + cmp.dialect.Quote(x.As.String()), nil
			} else {
				if strings.Contains(resStr, ".") {
					fieldName := strings.Split(resStr, ".")[1]
					fieldName = fieldName[1 : len(fieldName)-1]
					cmp.dict.ExprAlias[strings.ToLower(fieldName)] = types.OutputExpr{
						SqlNode: x,
						Expr: types.FiedlExpression{
							ExprContent: resStr,
						},
					} //resStr
				} else {
					if resStr == "?" {
						fmt.Println(resStr)
					}
					fieldName := resStr
					fieldName = fieldName[1 : len(fieldName)-1]
					cmp.dict.ExprAlias[strings.ToLower(fieldName)] = types.OutputExpr{
						SqlNode: x,
						Expr: types.FiedlExpression{
							ExprContent: resStr,
						},
					} //resStr
				}
				//cmp.dict.ExprAlias[strings.ToLower(strings.ToLower(resStr))] = resStr
				return resStr, nil
			}

		}
		tableAlias := x.As.String()
		field := ""

		if x.As.IsEmpty() {
			if len(cmp.dict.Tables) > 1 {
				return "", newCompilerError("table not foud", ERR_TABLE_NOT_FOUND)
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
				cmp.dict.ExprAlias[strings.ToLower(strings.ToLower(cmp.dict.StructField[matchField].Name))] = types.OutputExpr{
					SqlNode: x,
					Expr: types.FiedlExpression{
						ExprContent: retField,
						ExprType:    types.EXPR_TYPE(types.EXPR_TYPE_FIELD),
					},
				}
				return retField + " " + cmp.dialect.Quote(cmp.dict.StructField[matchField].Name), nil
			} else {
				if cmpType != C_SELECT {
					return cmp.dialect.Quote(tableAlias, field), nil
				}
				cmp.dict.ExprAlias[strings.ToLower(field)] = types.OutputExpr{
					SqlNode: x,
					Expr: types.FiedlExpression{
						ExprContent: cmp.dialect.Quote(tableAlias, field),
						ExprType:    types.EXPR_TYPE(types.EXPR_TYPE_FIELD),
					},
				}
				return cmp.dialect.Quote(tableAlias, field) + " " + cmp.dialect.Quote(field), nil
			}

		} else {
			expr, err := cmp.resolve(x.Expr, cmpType, args)
			if err != nil {
				return "", err
			}
			if x.As.IsEmpty() {
				return expr, nil
			} else {
				if cmpType != C_SELECT {
					return expr, nil
				}
				cmp.dict.ExprAlias[strings.ToLower(x.As.String())] = types.OutputExpr{
					SqlNode: x,
					Expr: types.FiedlExpression{
						ExprContent: expr,
					},
				}
				return expr + " " + cmp.dialect.Quote(x.As.String()), nil
			}

		}

	}
	if x, ok := expr.(*sqlparser.StarExpr); ok {
		if x.TableName.Qualifier.IsEmpty() {
			return "*", nil
		} else {
			return cmp.dialect.Quote(x.TableName.Qualifier.String()) + ".*", nil
		}

	}

	panic(fmt.Sprintf("%s Not support %T", expr, `compiler\cmp.resolve.go`))
}
