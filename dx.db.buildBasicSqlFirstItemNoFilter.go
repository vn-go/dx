package dx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/vn-go/dx/compiler"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
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
		// compiler, err := expr.CompileJoin(tableName, db.DB)
		// if err != nil {
		// 	return nil, err
		// }
		// tableName = compiler.Content
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

		if err != nil {
			return nil, err
		}
		// strField := compiler.Content

		sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(filterFields, ","), tableName)
		if filter != "" {

			sql += " WHERE " + filter
		}
		cmpInfo, err := compiler.Compile(sql, db.DriverName, false, false)
		if err != nil {
			return nil, err
		}
		sqlParse, err := dialect.BuildSql(cmpInfo.Info)
		if err != nil {
			return nil, err
		}
		return &basicSqlFirstItemNoFilterResult{
			sql:           sql,
			sqlCompiler:   sqlParse.Sql,
			keyFieldIndex: keyFieldIndex,
		}, nil
	})
	if err != nil {
		return "", "", nil, err
	}
	return ret.sql, ret.sqlCompiler, ret.keyFieldIndex, nil
}

type buildBasicSqlFirstItemNoFilterV2Key struct {
	typ reflect.Type
}

func (db *DB) buildBasicSqlFirstItemNoFilterV2(typ reflect.Type) (string, string, [][]int, error) {
	retType := reflect.TypeOf(basicSqlFirstItemNoFilterResult{})
	key := buildBasicSqlFirstItemNoFilterV2Key{
		typ: retType,
	}

	//key := typ.String() + "://" + db.DbName + "://" + db.DriverName
	ret, err := internal.OnceCall(key, func() (*basicSqlFirstItemNoFilterResult, error) {

		dialect := factory.DialectFactory.Create(db.DriverName)

		ent, err := model.ModelRegister.GetModelByType(typ)
		if err != nil {
			return nil, err
		}

		tableName := ent.Entity.TableName

		columns := ent.Entity.Cols

		fieldsSelect := make([]string, len(columns))
		//filterFields := []string{}
		keyFieldIndex := [][]int{}
		for i, col := range columns {
			if col.PKName != "" {
				//filterFields = append(filterFields, dialect.Quote(tableName, col.Name)+" =?")
				keyFieldIndex = append(keyFieldIndex, col.IndexOfField)
			}
			fieldsSelect[i] = dialect.Quote(tableName, col.Name) + " AS " + dialect.Quote(col.Field.Name)
		}
		//filter := strings.Join(filterFields, " AND ")
		linit := uint64(1)
		sqlInfo := &types.SqlInfo{
			StrSelect: strings.Join(fieldsSelect, ","),
			Limit:     &linit,
			From:      dialect.Quote(tableName),
		}

		sql, err := dialect.BuildSql(sqlInfo)
		if err != nil {
			return nil, err
		}

		return &basicSqlFirstItemNoFilterResult{
			sql: sql.Sql,
			//sqlCompiler:   filter,
			keyFieldIndex: keyFieldIndex,
		}, nil
		//return sql, compiler.Content, keyFieldIndex, nil
	})
	if err != nil {
		return "", "", nil, err
	}
	return ret.sql, ret.sqlCompiler, ret.keyFieldIndex, nil
}
