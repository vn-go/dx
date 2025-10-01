package postgres

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

func (d *postgresDialect) SqlFunction(delegator *types.DialectDelegateFunction) (string, error) {

	switch strings.ToLower(delegator.FuncName) {
	case "len":
		delegator.FuncName = "LENGTH"
		delegator.HandledByDialect = true
		return "LENGTH" + "(" + strings.Join(delegator.Args, ", ") + ")", nil
	case "concat":
		delegator.HandledByDialect = true
		castArgs := make([]string, len(delegator.Args))
		for i, x := range delegator.Args {
			if x[0] == '$' {
				castArgs[i] = x + "::text"
			} else {
				castArgs[i] = x
			}
		}
		return "CONCAT" + "(" + strings.Join(castArgs, ", ") + ")", nil
	default:
		if !d.isReleaseMode {
			panic(fmt.Sprintf("%s not implement at postgresDialect.SqlFunction, see %s", delegator.FuncName, `dialect\postgres\dialect.Postgres.SqlFunction.go`))
		}
		return "", nil
	}
}
