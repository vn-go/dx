package postgres

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
)

func (d *postgresDialect) SqlFunctionIf(delegator *types.DialectDelegateFunction) (string, error) {
	n := len(delegator.Args)
	stmts := []string{"CASE"}

	// Duyệt qua các cặp (condition, value)
	for i := 0; i+1 < n; i += 2 {
		stmts = append(stmts, fmt.Sprintf("WHEN %s THEN %s", delegator.Args[i], delegator.Args[i+1]))
	}

	// Nếu còn dư 1 đối số => ELSE
	if n%2 == 1 {
		stmts = append(stmts, fmt.Sprintf("ELSE %s", delegator.Args[n-1]))
	}

	stmts = append(stmts, "END")

	delegator.FuncName = "CASE"
	delegator.HandledByDialect = true

	return strings.Join(stmts, " "), nil
}
