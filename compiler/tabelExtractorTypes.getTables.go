package compiler

//type tabelExtractorTypes struct{}

// func (t *tabelExtractorTypes) getTablesNew(node sqlparser.SQLNode, visited map[string]bool, isSubQuery *bool) []string {
// 	if node == nil {
// 		return nil
// 	}

// 	ret := []string{}

// 	// Helper: thêm bảng nếu chưa tồn tại
// 	addTable := func(name string) {
// 		// if name != "" && !visited[name] {
// 		// 	visited[name] = true
// 		// 	ret = append(ret, name)
// 		// }
// 		ret = append(ret, name)
// 	}

// 	switch n := node.(type) {

// 	//----------------------------------------
// 	// SELECT
// 	//----------------------------------------
// 	case *sqlparser.Select:
// 		parts := []sqlparser.SQLNode{}
// 		//n.Having, n.From, n.Where, n.GroupBy, n.SelectExprs
// 		if n.Having != nil {
// 			parts = append(parts, n.Having)
// 		}
// 		if n.From != nil {
// 			parts = append(parts, n.From)
// 		}
// 		if n.Where != nil {
// 			parts = append(parts, n.Where)
// 		}
// 		if n.GroupBy != nil {
// 			parts = append(parts, n.GroupBy)
// 		}
// 		if n.SelectExprs != nil {
// 			parts = append(parts, n.SelectExprs)
// 		}
// 		for _, p := range parts {
// 			ret = append(ret, t.getTables(p, visited, isSubQuery)...)
// 		}

// 	//----------------------------------------
// 	// UNION
// 	//----------------------------------------
// 	case *sqlparser.Union:
// 		ret = append(ret, t.getTables(n.Left, visited, isSubQuery)...)
// 		ret = append(ret, t.getTables(n.Right, visited, isSubQuery)...)

// 	//----------------------------------------
// 	// SUBQUERY
// 	//----------------------------------------
// 	case *sqlparser.Subquery:
// 		*isSubQuery = true
// 		tbls := t.getTables(n.Select, visited, isSubQuery)
// 		ret = append(ret, tbls...)
// 		return ret

// 	//----------------------------------------
// 	// TABLES & ALIAS
// 	//----------------------------------------
// 	case sqlparser.TableExprs:
// 		checkIsSubQuery := false

// 		for _, e := range n {
// 			ret = append(ret, t.getTables(e, visited, &checkIsSubQuery)...)
// 		}
// 		*isSubQuery = (*isSubQuery) || checkIsSubQuery

// 	case *sqlparser.AliasedTableExpr:
// 		tables := t.getTables(n.Expr, visited, isSubQuery)
// 		if len(tables) > 0 {
// 			alias := n.As.String()
// 			for _, tbl := range tables {
// 				if alias == "" {
// 					addTable(tbl)
// 				} else {
// 					addTable(fmt.Sprintf("%s AS %s", tbl, alias))
// 				}
// 			}
// 		}

// 	case sqlparser.TableName:
// 		addTable(n.Name.String())

// 	case sqlparser.TableIdent:
// 		addTable(n.String())

// 	case *sqlparser.StarExpr:
// 		if !n.TableName.IsEmpty() {
// 			addTable(n.TableName.Name.String())
// 		}

// 	//----------------------------------------
// 	// JOIN
// 	//----------------------------------------
// 	case *sqlparser.JoinTableExpr:
// 		ret = append(ret, t.getTables(n.LeftExpr, visited, isSubQuery)...)
// 		ret = append(ret, t.getTables(n.RightExpr, visited, isSubQuery)...)
// 		ret = append(ret, t.getTables(n.Condition, visited, isSubQuery)...)

// 	case sqlparser.JoinCondition:
// 		ret = append(ret, t.getTables(n.On, visited, isSubQuery)...)

// 	//----------------------------------------
// 	// COLUMN
// 	//----------------------------------------
// 	case *sqlparser.ColName:
// 		if !n.Qualifier.IsEmpty() {
// 			addTable(n.Qualifier.Name.String())
// 		}

// 	//----------------------------------------
// 	// WHERE, GROUP BY, HAVING, EXPR, FUNCTION
// 	//----------------------------------------
// 	case *sqlparser.Where:
// 		if n != nil {
// 			ret = append(ret, t.getTables(n.Expr, visited, isSubQuery)...)
// 		} else {
// 			return ret
// 		}

// 	case sqlparser.GroupBy:
// 		for _, g := range n {
// 			ret = append(ret, t.getTables(g, visited, isSubQuery)...)
// 		}

// 	case *sqlparser.FuncExpr:
// 		for _, f := range n.Exprs {
// 			ret = append(ret, t.getTables(f, visited, isSubQuery)...)
// 		}

// 	case *sqlparser.AliasedExpr:
// 		ret = append(ret, t.getTables(n.Expr, visited, isSubQuery)...)

// 	case sqlparser.SelectExprs:
// 		for _, e := range n {
// 			ret = append(ret, t.getTables(e, visited, isSubQuery)...)
// 		}

// 	case *sqlparser.BinaryExpr:
// 		ret = append(ret, t.getTables(n.Left, visited, isSubQuery)...)
// 		ret = append(ret, t.getTables(n.Right, visited, isSubQuery)...)

// 	case *sqlparser.AndExpr:
// 		ret = append(ret, t.getTables(n.Left, visited, isSubQuery)...)
// 		ret = append(ret, t.getTables(n.Right, visited, isSubQuery)...)

// 	case *sqlparser.OrExpr:
// 		ret = append(ret, t.getTables(n.Left, visited, isSubQuery)...)
// 		ret = append(ret, t.getTables(n.Right, visited, isSubQuery)...)

// 	case *sqlparser.NotExpr:
// 		ret = append(ret, t.getTables(n.Expr, visited, isSubQuery)...)

// 	case *sqlparser.IsExpr:
// 		ret = append(ret, t.getTables(n.Expr, visited, isSubQuery)...)

// 	case *sqlparser.ParenExpr:
// 		ret = append(ret, t.getTables(n.Expr, visited, isSubQuery)...)

// 	//----------------------------------------
// 	// UPDATE
// 	//----------------------------------------
// 	case sqlparser.UpdateExprs:
// 		for _, e := range n {
// 			ret = append(ret, t.getTables(e, visited, isSubQuery)...)
// 		}

// 	case *sqlparser.UpdateExpr:
// 		ret = append(ret, t.getTables(n.Expr, visited, isSubQuery)...)
// 		if !n.Name.Qualifier.IsEmpty() {
// 			addTable(n.Name.Qualifier.Name.String())
// 		}
// 	case *sqlparser.ComparisonExpr:
// 		ret = append(ret, t.getTables(n.Left, visited, isSubQuery)...)
// 		ret = append(ret, t.getTables(n.Right, visited, isSubQuery)...)
// 	//----------------------------------------
// 	// IGNORE VALUES
// 	//----------------------------------------
// 	case *sqlparser.SQLVal:
// 		// Ignore literal value
// 		return nil

// 	default:
// 		panic(fmt.Sprintf("getTables: unhandled node type %T (%v)", n, n))
// 	}

// 	return ret
// }
