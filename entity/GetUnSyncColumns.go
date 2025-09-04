package entity

import (
	"strings"
	"sync"
)

type initGetUnSyncColumns struct {
	val  string
	once sync.Once
}

var cacheGetUnSyncColumns sync.Map

/*
this function will return a string of unsynced columns
unsynced columns are columns in Entity that do not exist in the db table
@dbColumnName is a list of column names in the db table
*/
func (e *Entity) GetUnSyncColumns(dbColumnName []string) string {
	actually, _ := cacheGetUnSyncColumns.LoadOrStore(e.EntityType, &initGetUnSyncColumns{})
	item := actually.(*initGetUnSyncColumns)
	item.once.Do(func() {
		dbCols := make(map[string]bool)
		for _, c := range dbColumnName {
			dbCols[c] = true
		}

		var unsync []string
		for _, col := range e.Cols {
			if _, found := dbCols[col.Name]; !found {
				unsync = append(unsync, col.Name)
			}
		}
		item.val = strings.Join(unsync, ", ")
	})
	return item.val

}
