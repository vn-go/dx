package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/sqlparser"
)

// select.union.go
func (s selectors) union(expr *sqlparser.Union, injector *injector) (*compilerResult, error) {
	sqlSource := []string{}
	injector.dict = newDictionary()
	l, err := s.selectStatement(expr.Left, injector, CMP_UNION)
	if err != nil {
		return nil, err
	}
	sqlSource = append(sqlSource, l.Content)
	sqlSource = append(sqlSource, expr.Type)
	injector.dict = newDictionary()
	r, err := s.selectStatement(expr.Right, injector, CMP_UNION)

	if err != nil {
		return nil, err
	}
	sqlSource = append(sqlSource, r.Content)
	l.Args = append(l.Args, r.Args...)
	l.Content = "\n " + strings.Join(sqlSource, "\n ")
	l.Fields.merge(r.Fields)
	return l, nil
}

func (s selectors) selectStatement(expr sqlparser.SelectStatement, injector *injector, cmpTyp CMP_TYP) (*compilerResult, error) {
	switch x := expr.(type) {
	case *sqlparser.Select:
		return selector.selects(x, injector, cmpTyp)
	case *sqlparser.Union:
		return selector.union(x, injector)
	default:
		panic(fmt.Sprintf("not implemented: %T. See selectors.selectStatement, %s", x, `sql\select.union.go`))
	}

}
