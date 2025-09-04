package entity

import "sync"

type initGetAutoValueColumns struct {
	once sync.Once
	val  []ColumnDef
}

func (e *Entity) GetAutoValueColumns() []ColumnDef {
	actual, _ := e.cacheGetAutoValueColumns.LoadOrStore(e.EntityType, &initGetAutoValueColumns{})
	init := actual.(*initGetAutoValueColumns)
	init.once.Do(func() {
		init.val = []ColumnDef{}
		for _, col := range e.Cols {
			if col.IsAuto {
				init.val = append(init.val, col)
			}
		}
	})
	return init.val
}
