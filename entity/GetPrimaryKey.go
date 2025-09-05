package entity

import "sync"

type initGetPrimaryKey struct {
	val  map[string][]ColumnDef
	once sync.Once
}

var cacheGetPrimaryKey sync.Map

/*
this function will return a map of primary key columns with their names
@return map[string][]ColumnDef //<-- map[Primary key constraint name][]ColumnDef
*/
func (e *Entity) GetPrimaryKey() map[string][]ColumnDef {
	actually, _ := cacheGetPrimaryKey.LoadOrStore(e.EntityType, &initGetPrimaryKey{})
	item := actually.(*initGetPrimaryKey)
	item.once.Do(func() {

		m := make(map[string][]ColumnDef)
		for _, col := range e.Cols {
			if col.PKName != "" {
				m[col.PKName] = append(m[col.PKName], col)

			}
		}
		item.val = m
	})
	return item.val

}
