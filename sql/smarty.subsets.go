package sql

import (
	sortList "sort"

	"github.com/vn-go/dx/sqlparser"
)

type subset struct {
}
type subsetInfo struct {
	query   string
	alias   string
	reIndex []int
}

func (s *subset) extractSubSetInfo(selectStm *sqlparser.Select, ret map[string]subsetInfo) (map[string]subsetInfo, error) {
	indexes := []int{}
	for i, x := range selectStm.SelectExprs {
		if aliasExpr, ok := x.(*sqlparser.AliasedExpr); ok {
			if fn := detect[*sqlparser.FuncExpr](aliasExpr.Expr); fn != nil {
				if fn.Name.Lowered() == "subsets" {
					selectExprs := fn.Exprs
					selector := &sqlparser.Select{
						SelectExprs: selectExprs,
					}
					query, reIndex, err := smartier.compile(selector, ret)
					if err != nil {
						return nil, err
					}
					ret[aliasExpr.As.Lowered()] = subsetInfo{

						alias:   aliasExpr.As.Lowered(),
						query:   query,
						reIndex: reIndex,
					}
					indexes = append(indexes, i)
				}
			}
		}
	}
	// remove subsets from selectExprs
	var exprs []sqlparser.SelectExpr = selectStm.SelectExprs

	sortList.Ints(indexes)

	// 2. Xóa từ CUỐI về ĐẦU
	for i := len(indexes) - 1; i >= 0; i-- {
		index := indexes[i]
		if index >= 0 && index < len(exprs) {
			// Kỹ thuật slicing để xóa
			exprs = append(exprs[:index], exprs[index+1:]...)
		}
	}
	selectStm.SelectExprs = exprs
	return ret, nil
}

var subsets = &subset{}
