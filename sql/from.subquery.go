package sql

import (
	"fmt"
	"strings"

	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/sqlparser"
)

// from.subquery.go
/*
	build subquery to compilerResult

	alias is alias of subquery
*/
func (f from) subquery(x *sqlparser.Subquery, alias string, injector *injector) (*compilerResult, error) {
	backupDict := injector.dict // backup dictionary

	defer func() {

		backupDict.fields = internal.UnionMap(backupDict.fields, injector.dict.fields)
		backupDict.tableAlias = internal.UnionMap(backupDict.tableAlias, injector.dict.tableAlias)
		backupDict.subqueryEntites = internal.UnionMap(backupDict.subqueryEntites, injector.dict.subqueryEntites)
		injector.dict = backupDict

	}()

	injector.dict = newDictionary() // sub query need new dictionary for compiling

	ret, err := f.selectStatement(x.Select, injector, CMP_SUBQUERY)
	if err != nil {
		return nil, err
	}
	backupDict.tableAlias[strings.ToLower(alias)] = alias
	if backupDict.subqueryEntites == nil {
		backupDict.subqueryEntites = make(map[string]subqueryEntity)
	}
	subQrEntity := subqueryEntity{
		fields: subqueryEntityFields{},
	}
	backupDict.subqueryEntites[strings.ToLower(alias)] = subQrEntity
	// backupDick.entities[strings.ToLower(alias)] = &entity.Entity{}
	if len(ret.selectedExprs) == 0 {
		err := fmt.Errorf("subquery has no fields")
		return nil, err
	}
	for _, x := range ret.selectedExprs {
		key := strings.ToLower(fmt.Sprintf("%s.%s", alias, x.Alias))
		backupDict.fields[key] = &dictionaryField{
			Expr:  injector.dialect.Quote(alias, x.Alias),
			Typ:   x.Typ,
			Alias: x.Alias,
		}
		subQrEntity.fields[key] = subqueryEntityField{
			source: alias,
			field:  x.Alias,
		}
	}
	ret.Content = fmt.Sprintf("(%s) %s", ret.Content, injector.dialect.Quote(alias))
	return ret, nil
}
