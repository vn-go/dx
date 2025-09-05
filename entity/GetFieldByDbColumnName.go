package entity

func (e *Entity) GetFieldByColumnName(colName string) string {
	col, ok := e.MapCols[colName]
	if ok {
		return col.Field.Name
	}
	return ""
}
