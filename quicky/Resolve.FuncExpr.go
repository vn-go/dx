package quicky

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

// Resolve.FuncExpr
func (r *Resolve) FuncExpr(n *sqlparser.FuncExpr, dialect types.Dialect, textParams []string, dynamicArgs []any, arg *ArgInspects, field *FieldInspects, dict *Dictionanry, cmpType C_TYPE) (string, error) {
	oldCmpType := cmpType
	cmpType = C_TYPE_FUNC
	defer func() { cmpType = oldCmpType }()
	argsItems := []string{}
	for _, x := range n.Exprs {
		content, err := r.Resolve(x, dialect, textParams, dynamicArgs, arg, field, dict, cmpType)
		if err != nil {
			return "", err
		}
		argsItems = append(argsItems, content)
	}

	delegetor := types.DialectDelegateFunction{
		FuncName: n.Name.String(),
		Args:     argsItems,
	}
	content, err := dialect.SqlFunction(&delegetor)
	if err != nil {
		return "", err
	}
	for _, x := range argsItems {
		if fx, ok := (*field)[x]; ok {
			fx.IsInAggregateFunc = delegetor.IsAggregate
			(*field)[x] = fx
		}
	}
	if delegetor.HandledByDialect {
		return content, nil
	} else {
		return fmt.Sprintf("%s(%s)", n.Name.String(), strings.Join(argsItems, ", ")), nil
	}
	panic("unimplemented")
}
