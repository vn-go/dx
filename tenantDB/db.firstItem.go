package tenantDB

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/factory"
	_ "github.com/vn-go/dx/dialect/mysql"
	"github.com/vn-go/dx/expr"
)

func buildBasicSqlFirstItemNoCache(typ reflect.Type, db *TenantDB, filter string) (string, error) {
	dialect := factory.DialectFactory.Create(db.GetDriverName())

	repoType, err := inserterObj.getEntityInfo(typ)
	if err != nil {
		return "", err
	}
	tableName := repoType.tableName
	compiler, err := expr.CompileJoin(tableName, db)
	if err != nil {
		return "", err
	}
	tableName = compiler.Content
	columns := repoType.entity.GetColumns()

	fieldsSelect := make([]string, len(columns))
	for i, col := range columns {
		fieldsSelect[i] = repoType.tableName + "." + col.Field.Name + " AS " + col.Field.Name
	}
	compiler.Context.Purpose = expr.BUILD_SELECT
	err = compiler.BuildSelectField(strings.Join(fieldsSelect, ", "))
	if err != nil {
		return "", err
	}
	strField := compiler.Content

	sql := fmt.Sprintf("SELECT %s FROM %s", strField, tableName)
	if filter != "" {
		compiler.Context.Purpose = expr.BUILD_WHERE //build_purpose_where
		err = compiler.BuildWhere(filter)
		if err != nil {
			return "", err
		}
		sql += " WHERE " + compiler.Content
	}
	sql = dialect.MakeSelectTop(sql, 1)
	return sql, nil
}

type initBuildBasicSqlFirstItem struct {
	once sync.Once
	val  string
	err  error
}

var cacheBuildBasicSqlFirstItem sync.Map

func BuildBasicSqlFirstItem(typ reflect.Type, db *TenantDB, filter string) (string, error) {
	key := db.GetDriverName() + "://" + db.GetDBName() + "/" + typ.String() + "/" + filter
	actual, _ := cacheBuildBasicSqlFirstItem.LoadOrStore(key, &initBuildBasicSqlFirstItem{})
	initBuild := actual.(*initBuildBasicSqlFirstItem)
	initBuild.once.Do(func() {
		sql, err := buildBasicSqlFirstItemNoCache(typ, db, filter)
		initBuild.val = sql
		initBuild.err = err
	})
	return initBuild.val, initBuild.err
}
