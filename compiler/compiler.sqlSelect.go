package compiler

import (
	"sync"
)

type SqlSelectInfo struct {
	StrSelect string

	StrWhere string

	Limit    *uint64
	Offset   *uint64
	StrOrder string

	From interface{} //<--string or SqlSelectInfo

	StrGroupBy string

	StrHaving string

	UnionNext *SqlSelectInfo
	UnionType string
}
type initCompileSql struct {
	val  *SqlSelectInfo
	err  error
	once sync.Once
}

var initCompileSqlCache sync.Map

// func CompileSql(sqlCmd, dbDriver string) (*SqlSelectInfo, error) {
// 	cmp, err := newBasicCompiler(sqlCmd, dbDriver)

// 	if err != nil {
// 		//init.err = err
// 		return nil, err
// 	}
// 	stmSelect := cmp.node.(*sqlparser.Select)
// 	visisted := make(map[string]bool)
// 	tableList := tabelExtractor.getTables(stmSelect.SelectExprs, visisted)
// 	tableList = append(tableList, tabelExtractor.getTables(stmSelect.From, visisted)...)
// 	tableList = append(tableList, tabelExtractor.getTables(stmSelect.Where, visisted)...)
// 	//cmp.initDict(stmSelect.SelectExprs)
// 	cmp.dict = cmp.CreateDictionary(tableList)

// }

// func CompileSql2(sqlCmd, dbDriver string) (*SqlSelectInfo, error) {
// 	key := fmt.Sprintf("%s@%s", sqlCmd, dbDriver)
// 	actually, _ := initCompileSqlCache.LoadOrStore(key, &initCompileSql{})
// 	init := actually.(*initCompileSql)
// 	init.once.Do(func() {

// 		init.val, init.err = cmp.resolveSelect(stmSelect.SelectExprs)
// 	})
// 	return init.val, init.err

// }
