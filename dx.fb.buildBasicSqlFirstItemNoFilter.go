package dx

import (
	"reflect"
	"sync"
)

type initBuildBasicSqlFirstItemNoFilter struct {
	once          sync.Once
	sqlSelect     string
	filter        string
	err           error
	keyFieldIndex [][]int
}

var cacheBuildBasicSqlFirstItemNoFilter sync.Map

func buildBasicSqlFirstItemNoFilter(typ reflect.Type, db *DB) (string, string, [][]int, error) {
	key := db.DriverName + "://" + db.DbName + "/" + typ.String()
	actual, _ := cacheBuildBasicSqlFirstItemNoFilter.LoadOrStore(key, &initBuildBasicSqlFirstItemNoFilter{})
	initBuild := actual.(*initBuildBasicSqlFirstItemNoFilter)
	initBuild.once.Do(func() {
		sql, filter, keyFieldIndex, err := buildBasicSqlFirstItemNoFilterNoCache(typ, db)
		initBuild.sqlSelect = sql
		initBuild.filter = filter
		initBuild.err = err
		initBuild.keyFieldIndex = keyFieldIndex
	})
	return initBuild.sqlSelect, initBuild.filter, initBuild.keyFieldIndex, initBuild.err
}
