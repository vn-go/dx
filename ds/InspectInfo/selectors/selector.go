package selectors

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/ds/errors"
	"github.com/vn-go/dx/ds/helper"
	"github.com/vn-go/dx/sqlparser"
)

type qualifierSelector struct {
	Name string
}

type selector struct {
	Qualifier *qualifierSelector
}

func (s *selector) NewQualifier(qualifier string) *qualifierSelector {
	return &qualifierSelector{
		Name: qualifier,
	}
}
func (s *selector) Resolve(nodes []any, injectInfo *common.InjectInfo, resoleInspectInfo common.ResolveQuery) (string, error) {
	strColsSelected := []string{}
	for _, node := range nodes {
		switch node := node.(type) {
		case *sqlparser.AliasedExpr:
			switch expr := node.Expr.(type) {
			case *sqlparser.FuncExpr:
				fnName := strings.ToLower(expr.Name.String())
				if injectInfo.Dict.AliasMap[fnName] != "" {
					result, err := s.NewQualifier(expr.Name.String()).SelectExprs(expr.Exprs, injectInfo)

					if err != nil {
						return "", err
					}

					strColsSelected = append(strColsSelected, result...)

				} else {
					result, err := s.FuncExpr(expr, injectInfo)
					if err != nil {
						return "", err
					}

					strColsSelected = append(strColsSelected, result)
				}
			case *sqlparser.ColName:
				r, err := s.ResolveColName(expr, injectInfo)
				if err != nil {
					return "", err
				}
				strColsSelected = append(strColsSelected, r.Content)
			default:
				panic(fmt.Sprintf("unimplemented node type: %T, see selector.Resolve, in file `%s`", expr, `ds\InspectInfo\selectors\selector.go`))
			}
		case *helper.InspectInfo:
			sql, err := resoleInspectInfo(node, injectInfo)
			if err != nil {
				return "", err
			}
			return sql.Sql, nil

		default:
			panic(fmt.Sprintf("unimplemented node type: %T, see selector.Resolve, in file `%s`", node, `ds\InspectInfo\selectors\selector.go`))
		}
	}
	return strings.Join(strColsSelected, ", "), nil
}

func (s *selector) ResolveSelectExprs(exprs sqlparser.SelectExprs, injectInfo *common.InjectInfo) ([]string, error) {
	strColsSelected := []string{}
	for _, x := range exprs {
		switch x := x.(type) {
		case *sqlparser.AliasedExpr:
			r, err := s.ResolveAliasedExpr(x, injectInfo)
			if err != nil {
				return nil, err
			}
			strColsSelected = append(strColsSelected, r.Content)
		default:
			panic(fmt.Sprintf("unimplemented node type: %T, see selector.ResolveSelectExprs, in file `%s`", x, `ds\InspectInfo\selectors\selector.go`))
		}

	}
	return strColsSelected, nil
}

func (s *selector) ResolveAliasedExpr(node *sqlparser.AliasedExpr, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	r, err := s.ResolveExpr(node.Expr, injectInfo)
	if err != nil {
		return nil, err
	}
	if node.As.IsEmpty() {
		return r, nil
	} else {
		r.Content = fmt.Sprintf("%s AS %s", r.Content, injectInfo.Dialect.Quote(node.As.String()))
		r.Content = fmt.Sprintf("%s AS %s", r.Content, node.As.String())
		return r, nil
	}

}

func (s *selector) ResolveExpr(expr sqlparser.Expr, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	switch expr := expr.(type) {
	case *sqlparser.ColName:
		return s.ResolveColName(expr, injectInfo)

	default:
		panic(fmt.Sprintf("unimplemented node type: %T, see selector.ResolveExpr, in file `%s`", expr, `ds\InspectInfo\selectors\selector.go`))
	}

}

func (s *selector) ResolveColName(expr *sqlparser.ColName, injectInfo *common.InjectInfo) (*common.ResolverContent, error) {
	key := strings.ToLower(fmt.Sprintf("%s.%s", expr.Qualifier.Name.String(), expr.Name.String()))
	if field, ok := injectInfo.Dict.FieldMap[key]; ok {
		originalContent := expr.Name.String()

		if injectInfo.SelectFields == nil {
			injectInfo.SelectFields = map[string]common.Expression{}
		}
		injectInfo.SelectFields[strings.ToLower(field.Content)] = common.Expression{
			Content:           field.Content,
			OriginalContent:   originalContent,
			Type:              common.EXPR_TYPE_FIELD,
			Alias:             field.Alias,
			IsInAggregateFunc: false,
		}

		return &common.ResolverContent{
			Content:         field.Content,
			OriginalContent: originalContent,
			AliasField:      field.Alias,
		}, nil
	} else {
		if !expr.Qualifier.IsEmpty() {
			return nil, errors.NewParseError("field `%s` not found in dataset '%s'", expr.Name.String(), expr.Qualifier.Name.String())
		} else {

			return nil, errors.NewParseError("field `%s` not found", expr.Name.String())
		}

	}
}

var Selector = &selector{}
