package entity

/*
this function will return a map of unique key columns with their names
@return map[string][]ColumnDef //<-- map[Unique constraint name][]ColumnDef
*/
func (e *Entity) GetUnique() map[string][]ColumnDef {
	m := make(map[string][]ColumnDef)
	for _, col := range e.Cols {
		if col.UniqueName != "" {
			m[col.UniqueName] = append(m[col.UniqueName], col)
		}
	}
	return m
}
