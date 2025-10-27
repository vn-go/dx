package sql

import (
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

type group struct {
}



func (g *group) resolve(expr sqlparser.GroupBy, injector *injector, reverse dictionaryFields) (*compilerResult, error) {
	itemsOfContent := []string{}
	itemsOfOriginalContent := []string{}
	ret := compilerResult{}
	for _, x := range expr {
		r, err := exp.resolve(x, injector, CMP_GROUP, reverse)
		if err != nil {
			return nil, err
		}
		itemsOfContent = append(itemsOfContent, r.Content)
		ret.selectedExprsReverse.merge(r.selectedExprsReverse)
		ret.Fields.merge(r.Fields)
		ret.nonAggregateFields.merge(r.nonAggregateFields)
		ret.selectedExprs.merge(r.selectedExprs)
		itemsOfOriginalContent = append(itemsOfOriginalContent, r.OriginalContent)
	}
	ret.Content = strings.Join(itemsOfContent, ", ")
	ret.OriginalContent = strings.Join(itemsOfOriginalContent, ", ")
	return &ret, nil
}

var groups = &group{}
