package sql

import "github.com/vn-go/dx/sqlparser"

//select.selects.go
func (s selectors) selects(expr *sqlparser.Select, injector *injector) (*compilerResult, error) {
	sql := sqlComplied{}

	r, err := froms.resolve(expr.From, injector)
	if err != nil {
		return nil, err
	}

	sql.source = r.Content
	r.Content = sql.String()
	return r, nil
}
