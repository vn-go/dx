package quicky

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/sqlparser"
)

type C_TYPE int

const (
	C_TYPE_FROM C_TYPE = iota
	C_TYPE_JOIN
	C_TYPE_SELECT
	C_TYPE_FUNC
)

type Resolve struct {
}

func (r *Resolve) Resolve(n sqlparser.SQLNode, dialect types.Dialect, textParams []string, dynamicArgs []any, arg *ArgInspects, field *FieldInspects, dict *Dictionanry, cmpType C_TYPE) (string, error) {
	switch n := n.(type) {
	case *sqlparser.AliasedExpr:
		ret, err := r.Resolve(n.Expr, dialect, textParams, dynamicArgs, arg, field, dict, cmpType)
		if err != nil {
			return "", err
		}
		if cmpType == C_TYPE_SELECT {
			if n.As.IsEmpty() {
				ret += " " + dialect.Quote((*field)[ret].Alias)
			} else {
				ret += " " + dialect.Quote(n.As.String())
			}

		}
		return ret, nil
	case *sqlparser.ComparisonExpr:
		return r.ComparisonExpr(n, dialect, textParams, dynamicArgs, arg, field, dict, cmpType)
	case *sqlparser.ColName:
		return r.ColName(n, dialect, textParams, dynamicArgs, arg, field, dict, cmpType)
	}
	panic(fmt.Sprintf("unhandled node type %T. See Resolve.Resolve '%s'", n, `quicky\resolve.go`))
}

//Resolve.ColName

var resolver = &Resolve{}
