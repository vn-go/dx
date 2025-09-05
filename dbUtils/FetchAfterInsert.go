package dbutils

import (
	"database/sql"
	"reflect"

	"github.com/vn-go/dx/entity"
)

func (r *inserter) fetchAfterInsert(res sql.Result, entity *entity.Entity, dataValue reflect.Value) error {
	// Nếu không có cột tự tăng thì bỏ qua
	autoCols := entity.GetAutoValueColumns()
	if len(autoCols) == 0 {
		return nil
	}

	lastID, err := res.LastInsertId() //<--loi khi chay voi mssql
	/*
		Cau query thuc the duoc goi den mssql nhu co dang sau
		 vi du
		 "LastInsertId is not supported. Please use the OUTPUT clause or add `select ID = convert(bigint, SCOPE_IDENTITY())` to the end of your query"  khi chay cau query sau
		 "INSERT INTO [departments] (
		 	[name],
			[code],
			[parent_id], ...) OUTPUT INSERTED.[id] VALUES (@p1, @p2, @p3, ...)"
	*/
	if err != nil {
		return err
	}

	for _, col := range autoCols {
		valField := dataValue.FieldByName(col.Field.Name)

		if valField.CanConvert(col.Field.Type) {
			valField.Set(reflect.ValueOf(lastID).Convert(col.Field.Type))

		}
	}

	return nil
}
