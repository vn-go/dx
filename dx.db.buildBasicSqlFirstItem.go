package dx

import (
	"reflect"
	"sync"
)

type initBuildBasicSqlFirstItem struct {
	once sync.Once
	val  string
	err  error
}

var cacheBuildBasicSqlFirstItem sync.Map

func (db *DB) buildBasicSqlFirstItem(typ reflect.Type, filter string) (string, error) {
	key := db.DriverName + "://" + db.DbName + "/" + typ.String() + "/" + filter
	actual, _ := cacheBuildBasicSqlFirstItem.LoadOrStore(key, &initBuildBasicSqlFirstItem{})
	initBuild := actual.(*initBuildBasicSqlFirstItem)
	initBuild.once.Do(func() {
		sql, err := db.buildBasicSqlFirstItemNoCache(typ, filter)
		initBuild.val = sql
		initBuild.err = err
	})
	return initBuild.val, initBuild.err
}
