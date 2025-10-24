package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// from.subquery.go
func (f from) subquery(x *sqlparser.Subquery, alias string, injector *injector) (*compilerResult, error) {
	backupDick := injector.dict

	defer func() {

		backupDick.fields = internal.UnionMap(backupDick.fields, injector.dict.fields)
		backupDick.tableAlias = internal.UnionMap(backupDick.tableAlias, injector.dict.tableAlias)
		injector.dict = backupDick

	}()

	injector.dict = newDictionary() // sub query need new dictionary for compiling
	ret, err := f.selectStatement(x.Select, injector)
	if err != nil {
		return nil, err
	}
	backupDick.tableAlias[strings.ToLower(alias)] = alias
	if backupDick.subqueryEntites == nil {
		backupDick.subqueryEntites = make(map[string]subqueryEntity)
	}
	backupDick.subqueryEntites[strings.ToLower(alias)] = subqueryEntity{}
	// backupDick.entities[strings.ToLower(alias)] = &entity.Entity{}
	for _, x := range ret.selectedExprs {
		key := strings.ToLower(fmt.Sprintf("%s.%s", alias, x.Alias))
		backupDick.fields[key] = &dictionaryField{
			Expr:  injector.dialect.Quote(alias, x.Alias),
			Typ:   x.Typ,
			Alias: x.Alias,
		}
	}
	ret.Content = fmt.Sprintf("(%s) %s", ret.Content, injector.dialect.Quote(alias))
	return ret, nil
}
