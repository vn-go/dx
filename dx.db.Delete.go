package dx

import (
	"fmt"
	"reflect"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/model"
)

type DeleteResult struct {
	RowsAffected int64
	Error        error
}

func (db *DB) Delete(item interface{}, filter string, args ...interface{}) DeleteResult {
	typ := reflect.TypeOf(item)
	dialect := factory.DialectFactory.Create(db.DriverName)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()

	}
	compiler, err := expr.NewExprCompiler(db.DB)
	if err != nil {
		return DeleteResult{Error: err}
	}
	model, err := model.ModelRegister.GetModelByType(typ) //db.getModelFromCache(typ)
	if err != nil {
		return DeleteResult{
			RowsAffected: 0,
			Error:        err,
		}
	}
	compiler.Context.Purpose = expr.BUILD_WHERE //build_purpose_where
	compiler.Context.Tables = []string{model.Entity.TableName}
	compiler.Context.Alias = map[string]string{model.Entity.TableName: model.Entity.TableName}
	compiler.Context.Dialect = dialect
	if filter == "" {
		return DeleteResult{Error: fmt.Errorf("filter is empty")}
	}
	err = compiler.BuildWhere(filter)
	if err != nil {
		return DeleteResult{Error: err}
	}
	filter = compiler.Content
	sql := "DELETE FROM " + dialect.Quote(model.Entity.TableName) + " WHERE " + filter
	r, err := db.Exec(sql, args...)
	if err != nil {
		return DeleteResult{Error: err}
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return DeleteResult{Error: err}
	}
	return DeleteResult{RowsAffected: rows}

}
