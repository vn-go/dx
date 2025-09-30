package dx

import (
	"fmt"
	"reflect"

	// "github.com/vn-go/dx/expr"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

type DeleteResult struct {
	RowsAffected int64
	Error        error
}
type deleteKey struct {
	//typ        reflect.Type
	tableName  string
	filter     string
	driverName string
}

func (db *DB) Delete(item interface{}, filter string, args ...interface{}) DeleteResult {
	typ := reflect.TypeOf(item)

	model, err := model.ModelRegister.GetModelByType(typ) //db.getModelFromCache(typ)
	if err != nil {
		return DeleteResult{
			Error: err,
		}
	}
	if filter == "" {
		return DeleteResult{Error: fmt.Errorf("filter is empty")}
	}
	key := deleteKey{
		tableName:  model.Entity.TableName,
		filter:     filter,
		driverName: db.DriverName,
	}
	//key := "Delete/" + "/" + db.DriverName + "/" + model.Entity.TableName + "/" + filter
	sqlExec, err := internal.OnceCall(key, func() (*types.SqlParse, error) {
		sql := "DELETE FROM " + model.Entity.TableName + " WHERE " + filter
		sqlInfo, err := compiler.Compile(sql, db.DriverName, false)

		if err != nil {
			return nil, err
		}
		dialect := factory.DialectFactory.Create(db.DriverName)
		return dialect.BuildSql(sqlInfo.Info)

	})

	if err != nil {
		return DeleteResult{
			Error: err,
		}
	}
	if Options.ShowSql {
		fmt.Println(sqlExec.Sql)
	}
	r, err := db.Exec(sqlExec.Sql, args...)
	if err != nil {
		return DeleteResult{Error: err}
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return DeleteResult{Error: err}
	}
	return DeleteResult{RowsAffected: rows}

}
