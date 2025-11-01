package sql

import "github.com/vn-go/dx/sqlparser"

type subset struct {
}
type subsetInfo struct {
	query   string
	alias   string
	reIndex []int
}

func (s *subset) extractSubSetInfo(selectStm *sqlparser.Select, ret map[string]subsetInfo) (map[string]subsetInfo, error) {

	for _, x := range selectStm.SelectExprs {
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

				}
			}
		}
	}
	return ret, nil
}

var subsets = &subset{}
