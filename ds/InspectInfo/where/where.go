package where

import (
	"fmt"

	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/ds/common"
	"github.com/vn-go/dx/ds/helper"
	"github.com/vn-go/dx/sqlparser"
)

type where struct {
}

func (w *where) Resolve(nodes []any, injectInfo *common.InjectInfo, query func(info *helper.InspectInfo, injectInfo *common.InjectInfo) (*types.SqlParse, error)) (string, error) {
	for _, node := range nodes {
		switch node.(type) {
		case *sqlparser.AliasedExpr:
			// TODO: resolve aliased expr

		default:
			panic(fmt.Sprintf("not support node type: %T. See where.Resolve, file %s", node, `ds\InspectInfo\where\where.go`))
		}
	}
	return "", nil
}

var Where = &where{}
