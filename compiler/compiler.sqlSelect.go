package compiler

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

// type initCompileSql struct {
// 	val  *SqlSelectInfo
// 	err  error
// 	once sync.Once
// }

// var initCompileSqlCache sync.Map
