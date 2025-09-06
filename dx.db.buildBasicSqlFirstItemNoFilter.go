package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/model"
)

type basicSqlFirstItemNoFilterResult struct {
	sql           string
	sqlCompiler   string
	keyFieldIndex [][]int
}

func (db *DB) buildBasicSqlFirstItemNoFilter(typ reflect.Type) (string, string, [][]int, error) {
	retType := reflect.TypeOf(basicSqlFirstItemNoFilterResult{})
	key := retType.String() + "@" + retType.PkgPath()

	//key := typ.String() + "://" + db.DbName + "://" + db.DriverName
	ret, err := internal.OnceCall(key, func() (*basicSqlFirstItemNoFilterResult, error) {

		dialect := factory.DialectFactory.Create(db.DriverName)

		repoType, err := model.ModelRegister.GetModelByType(typ)
		if err != nil {
			return nil, err
		}
		tableName := repoType.Entity.TableName
		compiler, err := expr.CompileJoin(tableName, db.DB)
		if err != nil {
			return nil, err
		}
		tableName = compiler.Content
		columns := repoType.Entity.Cols

		fieldsSelect := make([]string, len(columns))
		filterFields := []string{}
		keyFieldIndex := [][]int{}
		for i, col := range columns {
			if col.PKName != "" {
				filterFields = append(filterFields, repoType.Entity.TableName+"."+col.Name+" =?")
				keyFieldIndex = append(keyFieldIndex, col.IndexOfField)
			}
			fieldsSelect[i] = repoType.Entity.TableName + "." + col.Field.Name + " AS " + col.Field.Name
		}
		filter := strings.Join(filterFields, " AND ")
		compiler.Context.Purpose = expr.BUILD_SELECT //build_purpose_select
		err = compiler.BuildSelectField(strings.Join(fieldsSelect, ", "))
		if err != nil {
			return nil, err
		}
		strField := compiler.Content

		sql := fmt.Sprintf("SELECT %s FROM %s", strField, tableName)
		if filter != "" {
			compiler.Context.Purpose = expr.BUILD_WHERE //build_purpose_where
			err = compiler.BuildWhere(filter)
			if err != nil {
				return nil, err
			}

		}
		offset := uint64(1)
		sql = dialect.LimitAndOffset(sql, nil, &offset, "")
		return &basicSqlFirstItemNoFilterResult{
			sql:           sql,
			sqlCompiler:   compiler.Content,
			keyFieldIndex: keyFieldIndex,
		}, nil
		//return sql, compiler.Content, keyFieldIndex, nil
	})
	if err != nil {
		return "", "", nil, err
	}
	return ret.sql, ret.sqlCompiler, ret.keyFieldIndex, nil
}
