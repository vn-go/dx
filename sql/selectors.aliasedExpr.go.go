package sql

// selectors.aliasedExpr.go
// func (s selectors) aliasedExpr(expr *sqlparser.AliasedExpr, injector *injector) (*compilerResult, error) {
// 	switch t := expr.Expr.(type) {
// 	case *sqlparser.ColName:
// 		return s.colName(t, injector)
// 	case *sqlparser.BinaryExpr:

// 		ret, err := exp.resolve(t, injector)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if expr.As.IsEmpty() {
// 			return nil, newCompilerError("'%s' require alias", ret.OriginalContent)
// 		}

// 		ret.selectedExprs[strings.ToLower(ret.Content)] = &dictionaryField{
// 			Expr:              ret.Content,
// 			Typ:               -1,
// 			Alias:             expr.As.String(),
// 			IsInAggregateFunc: ret.IsInAggregateFunc,
// 		}
// 		ret.Content = ret.Content + " " + injector.dialect.Quote(expr.As.String())
// 		ret.selectedExprsReverse[strings.ToLower(expr.As.String())] = ret.selectedExprs[strings.ToLower(ret.Content)]

// 		return ret, nil
// 	case *sqlparser.FuncExpr:
// 		return exp.funcExpr(t, injector)
// 	default:
// 		panic(fmt.Sprintf("unimplemented: %T. See selectors.aliasedExpr, %s", t, `sql\selectors.aliasedExpr.go.go`))

// 	}

// }
