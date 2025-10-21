package common

import (
	"fmt"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

type resolver struct {
}

func (r *resolver) SQLVal(expr *sqlparser.SQLVal, injectInfo *InjectInfo) (*ResolverContent, error) {
	switch expr.Type {
	// case sqlparser.StrVal:
	// 	return &ResolverContent{
	// 		Content:         injectInfo.Dialect.ToParam(0, expr.Type),
	// 		OriginalContent: "'" + string(expr.Val) + "",
	// 	}, nil
	case sqlparser.IntVal:
		value, err := internal.Helper.ToIntFormBytes(expr.Val)
		if err != nil {
			return nil, err
		}
		injectInfo.Args = append(injectInfo.Args, ArgScaner{
			Value:      value,
			IsConstant: true,
		})
		return &ResolverContent{
			Content:         injectInfo.Dialect.ToParam(len(injectInfo.Args), expr.Type),
			OriginalContent: string(expr.Val),
		}, nil

	}
	panic(fmt.Sprintf("SQLVal not implemented for %s. See resolver.SQLVal,%s", expr.Type, `ds\common\resolver.go`))
}

var Resolver = &resolver{}
