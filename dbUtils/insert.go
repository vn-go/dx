package dbutils

import (
	"database/sql"
	"reflect"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/migate/migrator"
	"github.com/vn-go/dx/model"
)

type inserter struct {
}

func (r *inserter) Insert(db *db.DB, data interface{}) error {
	dialect := factory.DialectFactory.Create(db.Info.DriverName)
	dataValue := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		dataValue = dataValue.Elem()
	}
	modelInfo, err := model.ModelRegister.GetModelByType(typ)
	sqlText, args := dialect.MakeSqlInsert(modelInfo.Entity.TableName, modelInfo.Entity.Cols, data)
	sqlStmt, err := db.Prepare(sqlText)
	if err != nil {
		return err
	}
	defer sqlStmt.Close()

	switch dialect.Name() {
	case "mssql", "postgres":
		var insertedID int64
		err = sqlStmt.QueryRow(args...).Scan(&insertedID)
		if err == nil {
			err = r.fetchAfterInsertForQueryRow(modelInfo.Entity, dataValue, insertedID)
		}

	default:
		var res sql.Result
		res, err = sqlStmt.Exec(args...) // ✅ chỗ này đã hợp lệ
		if err == nil {
			err = r.fetchAfterInsert(res, modelInfo.Entity, dataValue)
		}
	}

	if err != nil {
		m, err1 := migrator.GetMigator(db)
		if err1 != nil {
			return err
		}
		schema, err1 := m.GetLoader().LoadFullSchema()
		if err1 != nil {
			return err
		}

		return r.parseDbError(schema, dialect, err, modelInfo.Entity)
	}

	return nil
}
