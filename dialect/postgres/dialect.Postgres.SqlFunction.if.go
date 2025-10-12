package postgres

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

func (d *postgresDialect) SqlFunctionIf(delegator *types.DialectDelegateFunction) (string, error) {
	n := len(delegator.Args)
	stmts := []string{"CASE"}
	args := []string{}
	isHasVar := false
	for i, x := range delegator.Args {
		if i == 0 {
			args = append(args, x)
			continue
		}
		if strings.Contains(x, "::") {
			args = append(args, strings.Split(x, "::")[0])
		} else {
			args = append(args, x)
			isHasVar = true
		}

	}
	if !isHasVar {
		args = delegator.Args
	}

	// Duyệt qua các cặp (condition, value)
	for i := 0; i+1 < n; i += 2 {
		stmts = append(stmts, fmt.Sprintf("WHEN %s THEN %s", args[i], args[i+1]))
	}

	// Nếu còn dư 1 đối số => ELSE
	if n%2 == 1 {
		stmts = append(stmts, fmt.Sprintf("ELSE %s", args[n-1]))
	}

	stmts = append(stmts, "END")

	delegator.FuncName = "CASE"
	delegator.HandledByDialect = true

	return strings.Join(stmts, " "), nil
}
