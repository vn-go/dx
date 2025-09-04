package entity

import "sync"

type initGetIndex struct {
	val  map[string][]ColumnDef
	once sync.Once
}

var cacheGetIndex sync.Map

/*
this function will return a map of index columns with their names
@return map[string][]ColumnDef //<-- map[Index constraint name][]ColumnDef
*/
func (e *Entity) GetIndex() map[string][]ColumnDef {
	actually, _ := cacheGetIndex.LoadOrStore(e.EntityType, &initGetIndex{})
	item := actually.(*initGetIndex)
	item.once.Do(func() {
		m := make(map[string][]ColumnDef)
		for _, col := range e.Cols {
			if col.IndexName != "" {
				m[col.IndexName] = append(m[col.IndexName], col)
			}
		}
		item.val = m
	})
	return item.val

}
