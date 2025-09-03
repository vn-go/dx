package tenantDB

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type initMakeScanInfo struct {
	once sync.Once
	info [][]int
}

var scanInfoCache sync.Map

func makeScanInfo(t reflect.Type, cols []string) [][]int {
	key := t.String() + strings.Join(cols, ",")
	actual, _ := scanInfoCache.LoadOrStore(key, &initMakeScanInfo{})
	initInfo := actual.(*initMakeScanInfo)
	initInfo.once.Do(func() {
		initInfo.info = makeScanInfoNoCache(t, cols)

	})
	return initInfo.info
}
func makeScanInfoNoCache(t reflect.Type, cols []string) [][]int {
	ret := make([][]int, len(cols))

	for i, col := range cols {
		field, ok := t.FieldByNameFunc(func(s string) bool {

			return strings.EqualFold(s, col)
		})
		if !ok {
			continue
		}
		ret[i] = field.Index
	}
	return ret
}

func doScan(rows *sql.Rows, dest interface{}) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to a struct or slice")
	}

	if destVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("only slice of struct is supported")
	}

	sliceVal := destVal.Elem()
	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("only slice of struct is supported")
	}

	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	scanInfo := makeScanInfo(elemType, cols)
	scanArgs := make([]interface{}, len(cols)) // reuse 1 lần
	dummies := make([]interface{}, len(cols))  // reuse dummy

	for i := range dummies {
		dummies[i] = new(interface{}) // allocate 1 lần cho mỗi dummy
	}

	for rows.Next() {
		elemVal := reflect.New(elemType).Elem()

		for i, index := range scanInfo {
			if len(index) > 0 {
				field := elemVal.FieldByIndex(index)
				if field.CanSet() {
					scanArgs[i] = field.Addr().Interface()
				} else {
					scanArgs[i] = dummies[i]
				}
			} else {
				scanArgs[i] = dummies[i]
			}
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}

		sliceVal.Set(reflect.Append(sliceVal, elemVal))
	}

	return nil
}

func (q *query) ToArray(items interface{}) error {
	sql, args := q.BuildSql()
	rows, err := q.db.Query(sql, args...)
	if err != nil {
		return err
	}
	//start := time.Now()
	//err = ToArrayUnsafeFast(rows, items)
	err = doScan(rows, items)
	//end := time.Now()
	//fmt.Println("ToArray time:", end.Sub(start).Milliseconds())

	if err != nil {
		return err
	}
	return nil

}
