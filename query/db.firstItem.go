package query

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vn-go/dx/dialect/factory"
	_ "github.com/vn-go/dx/dialect/mysql"
	"github.com/vn-go/dx/expr"
	"github.com/vn-go/dx/tenantDB"
)

func buildBasicSqlFirstItemNoCache(typ reflect.Type, db *tenantDB.TenantDB, filter string) (string, error) {
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
	tableName = compiler.content
	columns := repoType.entity.GetColumns()

	fieldsSelect := make([]string, len(columns))
	for i, col := range columns {
		fieldsSelect[i] = repoType.tableName + "." + col.Field.Name + " AS " + col.Field.Name
	}
	compiler.context.purpose = BUILD_SELECT
	err = compiler.buildSelectField(strings.Join(fieldsSelect, ", "))
	if err != nil {
		return "", err
	}
	strField := compiler.content

	sql := fmt.Sprintf("SELECT %s FROM %s", strField, tableName)
	if filter != "" {
		compiler.context.purpose = build_purpose_where
		err = compiler.buildWhere(filter)
		if err != nil {
			return "", err
		}
		sql += " WHERE " + compiler.content
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

func buildBasicSqlFirstItem(typ reflect.Type, db *tenantDB.TenantDB, filter string) (string, error) {
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
