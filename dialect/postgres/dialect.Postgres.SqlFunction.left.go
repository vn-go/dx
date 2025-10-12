package postgres

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
)

func (d *postgresDialect) sqlLeftFunc(delegator *types.DialectDelegateFunction) (string, error) {
	if len(delegator.Args) != 2 {
		return "", fmt.Errorf("%s rquire 2 arguments", delegator.FuncName)
	}
	delegator.HandledByDialect = true
	// p1 := delegator.Args[0]
	// p1 = strings.Split(p1, "::")[0] + "::citext"
	// p2 := delegator.Args[1]
	// p2 = strings.Split(p2, "::")[0] + "::int"
	ret := fmt.Sprintf("%s(%s,%s)", delegator.FuncName, delegator.Args[0], delegator.Args[1])
	//ret := fmt.Sprintf("substring(%s::text from 1 for %s )", delegator.Args[0], p2) //substring("T1"."username" from 1 for $1::int)
	return ret, nil
}
