package entity

/*
this function will return a map of unique key columns with their names
@return map[string][]ColumnDef //<-- map[Unique constraint name][]ColumnDef
*/
func (e *Entity) GetUnique() map[string]entityType {
	m := make(map[string]entityType)
	for _, col := range e.Cols {
		if col.UniqueName != "" {
			if _, ok := m[col.UniqueName]; !ok {
				m[col.UniqueName] = entityType{
					TableName: e.TableName,
					Cols:      []ColumnDef{},
				}
			}
			d := m[col.UniqueName]
			d.Cols = append(d.Cols, col)
			m[col.UniqueName] = d
		}
	}
	ret := make(map[string]entityType)
	for k, v := range m {
		ret["UQ_"+v.TableName+"__"+k] = v
	}
	return ret
}
