package quicky

import (
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

func (r *Resolve) ComparisonExpr(n *sqlparser.ComparisonExpr, dialect types.Dialect, textParams []string, dynamicArgs []any, arg *ArgInspects, field *FieldInspects, dict *Dictionanry, cmpType C_TYPE) (string, error) {
	left, err := r.Resolve(n.Left, dialect, textParams, dynamicArgs, arg, field, dict, cmpType)
	if err != nil {
		return "", err
	}
	right, err := r.Resolve(n.Right, dialect, textParams, dynamicArgs, arg, field, dict, cmpType)
	if err != nil {
		return "", err
	}
	return left + " " + n.Operator + " " + right, nil
}
