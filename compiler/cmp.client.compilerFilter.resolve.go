package compiler

// import (
// 	"fmt"
// 	"strconv"
// 	"strings"

// 	"github.com/vn-go/dx/dialect/types"
// 	"github.com/vn-go/dx/internal"
// 	"github.com/vn-go/dx/sqlparser"
// )

// // func (cmp *compilerFilterType) Resolvev1(dialect types.Dialect, strFilter string, fields map[string]string, n sqlparser.SQLNode) (*CompilerFilterTypeResult, error) {
// // 	// Comparison ( =, >, <, >=, <=, <> ...)
// // 	if x, ok := n.(*sqlparser.ComparisonExpr); ok {
// // 		if _, ok := x.Left.(*sqlparser.SQLVal); ok {
// // 			return nil, NewCompilerError(fmt.Sprintf(
// // 				"'%s' is invalid: left side of comparison must be a field or expression", strFilter))
// // 		}

// // 		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
// // 		if err != nil {
// // 			return nil, err
// // 		}

// // 		if left.expr == right.expr || left.fieldExpr == right.fieldExpr {
// // 			return nil, NewCompilerError(fmt.Sprintf(
// // 				"'%s' is invalid: comparison between identical expressions", strFilter))
// // 		}

// // 		expr := left.expr + " " + x.Operator + " " + right.expr
// // 		fieldExpr := left.fieldExpr + " " + x.Operator + " " + right.fieldExpr

// // 		fieldsSelected := mergeFieldMaps(left.fields, right.fields)
// // 		return &CompilerFilterTypeResult{expr: expr, fieldExpr: fieldExpr, fields: fieldsSelected}, nil
// // 	}

// // 	// Binary (bitwise hoặc các phép toán nhị phân)
// // 	if x, ok := n.(*sqlparser.BinaryExpr); ok {
// // 		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
// // 		if err != nil {
// // 			return nil, err
// // 		}

// // 		expr := left.expr + " " + x.Operator + " " + right.expr
// // 		fieldsSelected := mergeFieldMaps(left.fields, right.fields)
// // 		return &CompilerFilterTypeResult{expr: expr, fields: fieldsSelected}, nil
// // 	}

// // 	// Column name
// // 	if x, ok := n.(*sqlparser.ColName); ok {
// // 		name := strings.ToLower(x.Name.String())
// // 		if name == "yes" || name == "no" || name == "true" || name == "false" {
// // 			return &CompilerFilterTypeResult{expr: dialect.ToBool(name)}, nil
// // 		}

// // 		if v, ok := fields[name]; ok {
// // 			return &CompilerFilterTypeResult{
// // 				expr:      v,
// // 				fieldExpr: dialect.Quote(x.Name.String()),
// // 				fields:    map[string]string{name: x.Name.String()},
// // 			}, nil
// // 		}

// // 		validFields := make([]string, 0, len(fields))
// // 		for k := range fields {
// // 			validFields = append(validFields, k)
// // 		}
// // 		return nil, newCompilerError(fmt.Sprintf(
// // 			"Unknown field '%s'. Valid fields: [%s]. Expression: '%s'",
// // 			x.Name.String(), strings.Join(validFields, ","), strFilter), ERR)
// // 	}

// // 	// Constant or Parameter
// // 	if x, ok := n.(*sqlparser.SQLVal); ok {
// // 		v := string(x.Val)
// // 		if strings.HasPrefix(v, ":v") {
// // 			return &CompilerFilterTypeResult{expr: "?", fieldExpr: "?", isConstant: true}, nil
// // 		}
// // 		switch {
// // 		case x.Type == sqlparser.StrVal || internal.Helper.IsString(v):
// // 			return &CompilerFilterTypeResult{expr: dialect.ToText(v), fieldExpr: dialect.ToText(v), isConstant: true}, nil
// // 		case internal.Helper.IsBool(v):
// // 			return &CompilerFilterTypeResult{expr: dialect.ToBool(v), fieldExpr: dialect.ToBool(v), isConstant: true}, nil
// // 		case internal.Helper.IsFloatNumber(v), internal.Helper.IsNumber(v):
// // 			return &CompilerFilterTypeResult{expr: v, fieldExpr: v, isConstant: true}, nil
// // 		default:
// // 			return nil, NewCompilerError(fmt.Sprintf("Invalid constant '%s' in expression '%s'", v, strFilter))
// // 		}
// // 	}

// // 	// Logical AND
// // 	if x, ok := n.(*sqlparser.AndExpr); ok {
// // 		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		if left.isConstant || right.isConstant {
// // 			return nil, NewCompilerError(fmt.Sprintf(
// // 				"'%s' is invalid: operator AND requires both sides to be field or expression", strFilter))
// // 		}

// // 		fieldsSelected := mergeFieldMaps(left.fields, right.fields)
// // 		return &CompilerFilterTypeResult{
// // 			expr:      left.expr + " AND " + right.expr,
// // 			fieldExpr: left.fieldExpr + " AND " + right.fieldExpr,
// // 			fields:    fieldsSelected,
// // 		}, nil
// // 	}

// // 	// Logical OR
// // 	if x, ok := n.(*sqlparser.OrExpr); ok {
// // 		left, err := cmp.Resolve(dialect, strFilter, fields, x.Left)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		right, err := cmp.Resolve(dialect, strFilter, fields, x.Right)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		if left.isConstant || right.isConstant {
// // 			return nil, NewCompilerError(fmt.Sprintf(
// // 				"'%s' is invalid: operator OR requires both sides to be field or expression", strFilter))
// // 		}

// // 		fieldsSelected := mergeFieldMaps(left.fields, right.fields)
// // 		return &CompilerFilterTypeResult{
// // 			expr:      left.expr + " OR " + right.expr,
// // 			fieldExpr: left.fieldExpr + " OR " + right.fieldExpr,
// // 			fields:    fieldsSelected,
// // 		}, nil
// // 	}

// // 	// Logical NOT
// // 	if x, ok := n.(*sqlparser.NotExpr); ok {
// // 		left, err := cmp.Resolve(dialect, strFilter, fields, x.Expr)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		fieldsSelected := map[string]string{}
// // 		if left.fields != nil {
// // 			for k, v := range left.fields {
// // 				fieldsSelected[k] = v
// // 			}
// // 		}
// // 		return &CompilerFilterTypeResult{
// // 			expr:      "NOT " + left.expr,
// // 			fieldExpr: "NOT " + left.fieldExpr,
// // 			fields:    fieldsSelected,
// // 		}, nil
// // 	}

// // 	// Function
// // 	if x, ok := n.(*sqlparser.FuncExpr); ok {
// // 		return cmp.ResolveFunc(dialect, strFilter, fields, x)
// // 	}

// // 	// Aliased expression
// // 	if x, ok := n.(*sqlparser.AliasedExpr); ok {
// // 		return cmp.Resolve(dialect, strFilter, fields, x.Expr)
// // 	}

// // 	if isDebugMode {
// // 		panic(fmt.Sprintf("Not implemented type %T in Resolve() of compilerFilter.go", n))
// // 	}
// // 	return nil, newCompilerError(fmt.Sprintf("'%s' is invalid expression", strFilter), ERR)
// // }

// // Helper to merge selected fields
// func mergeFieldMaps(a, b map[string]string) map[string]string {
// 	m := map[string]string{}
// 	for k, v := range a {
// 		m[k] = v
// 	}
// 	for k, v := range b {
// 		if _, ok := m[k]; !ok {
// 			m[k] = v
// 		}
// 	}
// 	return m
// }
// func (s *compilerFilterType) Resolve(n sqlparser.SQLNode, strFilter string, fields map[string]string, dialect types.Dialect, argIndex *int) (*CompilerFilterTypeResult, error) {
// 	if n == nil {
// 		return nil, nil
// 	}

// 	switch x := n.(type) {

// 	// ---- Identifier / Field ----
// 	case *sqlparser.ColName:
// 		name := x.Name.String()
// 		mapped, ok := fields[name]
// 		if !ok {
// 			return nil, NewCompilerError(fmt.Sprintf("Unknown field '%s' in '%s'", name, strFilter))
// 		}

// 		return &CompilerFilterTypeResult{
// 			Expr:       dialect.Quote(mapped),
// 			FieldExpr:  dialect.Quote(mapped),
// 			Fields:     map[string]string{mapped: name},
// 			IsConstant: false,
// 		}, nil

// 	// ---- Constant values ----
// 	case *sqlparser.SQLVal:
// 		v := string(x.Val)

// 		// nếu giá trị bắt đầu bằng :v (tham số động)
// 		if strings.HasPrefix(v, ":v") {
// 			*argIndex++
// 			placeholder := "?"
// 			return &CompilerFilterTypeResult{
// 				Expr:       placeholder,
// 				FieldExpr:  placeholder,
// 				IsConstant: true,
// 				Args:       []interface{}{nil}, // giá trị sẽ được bind sau
// 			}, nil
// 		}

// 		// parse literal
// 		var val interface{}
// 		switch {
// 		case internal.Helper.IsBool(v):
// 			val = internal.Helper.ToBool(v)
// 		case internal.Helper.IsFloatNumber(v):
// 			f, _ := strconv.ParseFloat(v, 64)
// 			val = f
// 		case internal.Helper.IsNumber(v):
// 			i, _ := strconv.ParseInt(v, 10, 64)
// 			val = i
// 		case internal.Helper.IsString(v):
// 			val = internal.Helper.TrimStringLiteral(v)
// 		default:
// 			return nil, NewCompilerError(fmt.Sprintf("'%s' in '%s' is invalid value", v, strFilter))
// 		}

// 		*argIndex++
// 		placeholder := "?"
// 		return &CompilerFilterTypeResult{
// 			Expr:       placeholder,
// 			FieldExpr:  placeholder,
// 			IsConstant: true,
// 			Args:       []interface{}{val},
// 		}, nil

// 	// ---- Comparison: =, >, <, etc. ----
// 	case *sqlparser.ComparisonExpr:
// 		left, err := s.Resolve(x.Left, strFilter, fields, dialect, argIndex)
// 		if err != nil {
// 			return nil, err
// 		}
// 		right, err := s.Resolve(x.Right, strFilter, fields, dialect, argIndex)
// 		if err != nil {
// 			return nil, err
// 		}

// 		expr := fmt.Sprintf("%s %s %s", left.Expr, x.Operator, right.Expr)
// 		fieldExpr := fmt.Sprintf("%s %s %s", left.FieldExpr, x.Operator, right.FieldExpr)
// 		args := append(append([]interface{}{}, left.Args...), right.Args...)

// 		return &CompilerFilterTypeResult{
// 			Expr:      expr,
// 			FieldExpr: fieldExpr,
// 			Fields:    mergeFields(left.Fields, right.Fields),
// 			Args:      args,
// 		}, nil

// 	// ---- Logical operators ----
// 	case *sqlparser.AndExpr:
// 		left, err := s.Resolve(x.Left, strFilter, fields, dialect, argIndex)
// 		if err != nil {
// 			return nil, err
// 		}
// 		right, err := s.Resolve(x.Right, strFilter, fields, dialect, argIndex)
// 		if err != nil {
// 			return nil, err
// 		}
// 		expr := fmt.Sprintf("(%s AND %s)", left.Expr, right.Expr)
// 		fieldExpr := fmt.Sprintf("(%s AND %s)", left.FieldExpr, right.FieldExpr)
// 		args := append(append([]interface{}{}, left.Args...), right.Args...)

// 		return &CompilerFilterTypeResult{
// 			Expr:      expr,
// 			FieldExpr: fieldExpr,
// 			Fields:    mergeFields(left.Fields, right.Fields),
// 			Args:      args,
// 		}, nil

// 	case *sqlparser.OrExpr:
// 		left, err := s.Resolve(x.Left, strFilter, fields, dialect, argIndex)
// 		if err != nil {
// 			return nil, err
// 		}
// 		right, err := s.Resolve(x.Right, strFilter, fields, dialect, argIndex)
// 		if err != nil {
// 			return nil, err
// 		}
// 		expr := fmt.Sprintf("(%s OR %s)", left.Expr, right.Expr)
// 		fieldExpr := fmt.Sprintf("(%s OR %s)", left.FieldExpr, right.FieldExpr)
// 		args := append(append([]interface{}{}, left.Args...), right.Args...)

// 		return &CompilerFilterTypeResult{
// 			Expr:      expr,
// 			FieldExpr: fieldExpr,
// 			Fields:    mergeFields(left.Fields, right.Fields),
// 			Args:      args,
// 		}, nil

// 	// ---- NOT operator ----
// 	case *sqlparser.NotExpr:
// 		inner, err := s.Resolve(x.Expr, strFilter, fields, dialect, argIndex)
// 		if err != nil {
// 			return nil, err
// 		}
// 		expr := fmt.Sprintf("(NOT %s)", inner.Expr)
// 		fieldExpr := fmt.Sprintf("(NOT %s)", inner.FieldExpr)
// 		return &CompilerFilterTypeResult{
// 			Expr:      expr,
// 			FieldExpr: fieldExpr,
// 			Fields:    inner.Fields,
// 			Args:      inner.Args,
// 		}, nil

// 	// ---- Function calls ----
// 	case *sqlparser.FuncExpr:
// 		fname := strings.ToUpper(x.Name.String())
// 		if !dialect.IsSafeFunc(fname) {
// 			return nil, NewCompilerError(fmt.Sprintf("Function '%s' is not allowed in '%s'", fname, strFilter))
// 		}

// 		args := make([]*CompilerFilterTypeResult, 0)
// 		mergedArgs := make([]interface{}, 0)
// 		for _, a := range x.Exprs {
// 			ae, err := s.Resolve(a, strFilter, fields, dialect, argIndex)
// 			if err != nil {
// 				return nil, err
// 			}
// 			args = append(args, ae)
// 			mergedArgs = append(mergedArgs, ae.Args...)
// 		}

// 		argExprs := make([]string, len(args))
// 		for i, a := range args {
// 			argExprs[i] = a.Expr
// 		}

// 		expr := fmt.Sprintf("%s(%s)", fname, strings.Join(argExprs, ", "))
// 		return &CompilerFilterTypeResult{
// 			Expr:      expr,
// 			FieldExpr: expr,
// 			Fields:    mergeAllFields(args),
// 			Args:      mergedArgs,
// 		}, nil

// 	default:
// 		return nil, NewCompilerError(fmt.Sprintf("Unsupported expression in '%s'", strFilter))
// 	}
// }
// func mergeFields(a, b map[string]string) map[string]string {
// 	out := make(map[string]string, len(a)+len(b))
// 	for k, v := range a {
// 		out[k] = v
// 	}
// 	for k, v := range b {
// 		out[k] = v
// 	}
// 	return out
// }

// func mergeAllFields(list []*CompilerFilterTypeResult) map[string]string {
// 	out := make(map[string]string)
// 	for _, x := range list {
// 		for k, v := range x.Fields {
// 			out[k] = v
// 		}
// 	}
// 	return out
// }
// func TrimStringLiteral(s string) string {
// 	if len(s) == 0 {
// 		return s
// 	}
// 	if (s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"') {
// 		s = s[1 : len(s)-1]
// 	}
// 	// Chuẩn hóa escape ký tự nháy đơn trong SQL
// 	s = strings.ReplaceAll(s, "''", "'")
// 	s = strings.ReplaceAll(s, `\"`, `"`)
// 	return s
// }
