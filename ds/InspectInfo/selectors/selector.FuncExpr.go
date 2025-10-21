package selectors

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/ds/helper"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// selector.FuncExpr.go
func (s *selector) FuncExpr(expr *sqlparser.FuncExpr, injectInfo *common.InjectInfo) (string, error) {
	backupSelecteFields := injectInfo.SelectFields
	defer func() {
		injectInfo.SelectFields = internal.UnionMap(injectInfo.SelectFields, backupSelecteFields)
	}()
	injectInfo.SelectFields = map[string]common.Expression{}
	fName := expr.Name.String()
	if fName == helper.GET_PARAMS_FUNC {
		panic(fmt.Sprintf("unimplemented: %s, see:selector.FuncExpr.GET_PARAMS_FUNC ", fName))
	}
	var r *common.ResolverContent
	var err error
	delegator := &types.DialectDelegateFunction{
		FuncName: fName,
	}
	for _, arg := range expr.Exprs {

		r, err = s.SelectExpr(arg, injectInfo)
		if err != nil {
			return "", err
		}
		delegator.Args = append(delegator.Args, r.Content)
		delegator.ArgTypes = append(delegator.ArgTypes, r.DbTYpe)
	}
	for k, v := range injectInfo.SelectFields {
		v.IsInAggregateFunc = delegator.IsAggregate
		injectInfo.SelectFields[k] = v
	}

	strRet, err := injectInfo.Dialect.SqlFunction(delegator)
	if err != nil {
		return "", err
	}
	if delegator.HandledByDialect {
		return strRet, nil
	} else {
		return fmt.Sprintf("%s(%s)", fName, strings.Join(delegator.Args, ",")), nil
	}
}
