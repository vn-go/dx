package expr

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vn-go/dx/db"
	"github.com/vn-go/dx/dialect/factory"
	"github.com/vn-go/dx/dialect/types"
	"github.com/vn-go/dx/internal"
	"github.com/vn-go/dx/migate/migrator"
	"github.com/vn-go/dx/sqlparser"
	// "github.com/vn-go/dx/sqlparser"
)

type build_purpose int

const (
	build_purpose_select build_purpose = iota
	build_purpose_join
	build_purpose_where
	build_purpose_group
	build_purpose_having
	build_purpose_order
	build_purpose_limit
	build_purpose_offset
	build_purpose_for_function
	build_purpose_for_update
)

type exprCompileContext struct {
	tables []string
	/*
		The purpose of this field is track table name is already in database
	*/
	schema           *map[string]bool
	alias            map[string]string
	joinAlias        map[string]string
	aliasToDbTable   map[string]string
	dialect          types.Dialect
	purpose          build_purpose
	stackAliasFields internal.Stack[string]
	stackAliasTables internal.Stack[string]
	paramIndex       int
}

func (e *exprCompileContext) pluralTableName(tableName string) string {
	if e.schema != nil {
		if _, ok := (*e.schema)[tableName]; ok {
			return tableName
		} else {
			if _, ok := e.alias[tableName]; ok {
				return tableName
			} else {
				return internal.Utils.Pluralize(tableName)
			}
		}
	} else {
		if _, ok := e.alias[tableName]; ok {
			return tableName
		} else {
			return internal.Utils.Pluralize(tableName)
		}
	}
}

type exprCompiler struct {
	context *exprCompileContext
	content string
}

func (e *exprCompiler) buildSortField(selector string) error {
	e.context.purpose = build_purpose_order
	selector = internal.Helper.QuoteExpression(selector)
	sqlTest := "select tmp order by " + selector
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlSelect, ok := stm.(*sqlparser.Select); ok {

		ret, err := exprs.compile(e.context, sqlSelect.OrderBy)
		if err != nil {
			return err
		}
		e.content = ret

	}

	return nil
}
func (e *exprCompiler) buildSelectField(selector string) error {
	e.context.purpose = build_purpose_select
	selector = internal.Helper.QuoteExpression(selector)
	sqlTest := "select " + selector
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlSelect, ok := stm.(*sqlparser.Select); ok {
		selectors := make([]string, len(sqlSelect.SelectExprs))
		for i, expr := range sqlSelect.SelectExprs {
			if sqlExpr, ok := expr.(*sqlparser.AliasedExpr); ok {
				if !sqlExpr.As.IsEmpty() {
					e.context.stackAliasFields.Push(sqlExpr.As.String())
				}
				if sqlExpr.Expr != nil {
					strResult, err := exprs.compile(e.context, sqlExpr.Expr)

					if err != nil {
						return err
					}
					selectors[i] = strResult

				}
			} else {
				panic(fmt.Errorf("unsupported select type is %T", expr))
			}
		}
		e.content = strings.Join(selectors, ", ")
	}

	return nil
}
func (e *exprCompiler) buildSetter(stterExpr string) error {
	stterExpr = internal.Helper.QuoteExpression(stterExpr)

	sqlTest := "update test set " + stterExpr
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlUpdate, ok := stm.(*sqlparser.Update); ok {
		strResults := []string{}
		for _, expr := range sqlUpdate.Exprs {
			strResult, err := exprs.compile(e.context, expr)
			if err != nil {
				return err
			}
			strResults = append(strResults, strResult)

		}
		e.content = strings.Join(strResults, ", ")
	}

	return nil
}
func (e *exprCompiler) build(joinText string) error {
	joinText = internal.Helper.QuoteExpression(joinText)

	sqlTest := "select * from " + joinText
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlSelect, ok := stm.(*sqlparser.Select); ok {

		for _, expr := range sqlSelect.From {
			strResult, err := exprs.compile(e.context, expr)
			if err != nil {
				return err
			}
			e.content = strResult
		}
	}

	return nil

}
func (e *exprCompiler) buildWhere(where string) error {
	where = internal.Helper.QuoteExpression(where)
	e.context.purpose = build_purpose_where

	sqlTest := "select * from tmp where" + where
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlSelect, ok := stm.(*sqlparser.Select); ok {
		strResult, err := exprs.compile(e.context, sqlSelect.Where.Expr)
		// for _, expr := range sqlSelect.From {
		// 	strResult, err := exprs.compile(e.context, expr)
		if err != nil {
			return err
		}
		e.content = strResult
	}

	return nil

}

type initNewExprCompiler struct {
	once sync.Once
	val  *cacheNewExprCompilerItem
	err  error
}
type cacheNewExprCompilerItem struct {
	schema  map[string]bool
	dialect types.Dialect
}

var exprCompilerCache sync.Map

func NewExprCompiler(db *db.DB) (*exprCompiler, error) {
	key := db.Info.DriverName + "://" + db.Info.DbName
	actual, _ := exprCompilerCache.LoadOrStore(key, &initNewExprCompiler{})

	init := actual.(*initNewExprCompiler)
	init.once.Do(func() { //<-- thiet lap cau hinh bien dich
		init.val = &cacheNewExprCompilerItem{}
		m, err := migrator.GetMigator(db) //<-- make sure all table was migrated

		if err != nil {
			init.err = err
			return
		}

		err = m.DoMigrates() //<-- thuc hien migrate
		if err != nil {
			init.err = err //<-- loi bo bien dich khoi dong bi hong
			return
		}
		loader := m.GetLoader()              //<-- khoi tao bo loader cua migrate
		tables, err := loader.LoadAllTable() // <--- lay danh sach cac bang trong database va danh sach cac ban da duoc migrate
		if err != nil {
			init.err = err
			return
		}
		dialect := factory.DialectFactory.Create(db.Info.DriverName) //<-- khoi tao dialect, neu kg co dialect se kg the bien dich cua phap dung
		schema := map[string]bool{}
		for k, _ := range tables {
			schema[k] = true
		}
		init.val.schema = schema   //<-- bo bien dich can danh sach cac bang trong database
		init.val.dialect = dialect //<-- bo bien dich can dialect de quote va parse parameter,
		//cung nhu xua ly cac ham theo tu RDMMS rieng  biet vi du voi MSSQL la LEN khi su dung voi PostgreSQL la LENGTH,
		//haoc DAY trong SQLServer -> Postgres la DateExtract
	})
	if init.err != nil {
		return nil, init.err
	}

	ret := &exprCompiler{
		context: &exprCompileContext{
			tables:           make([]string, 0),
			schema:           &init.val.schema,
			alias:            make(map[string]string),
			aliasToDbTable:   make(map[string]string),
			dialect:          init.val.dialect,
			purpose:          build_purpose_select,
			stackAliasFields: internal.Stack[string]{},
			stackAliasTables: internal.Stack[string]{},
		},
	}
	return ret, nil
}
func CompileJoin(joinText string, db *db.DB) (*exprCompiler, error) {
	compiler, err := NewExprCompiler(db)
	if err != nil {
		return nil, err
	}
	compiler.context.purpose = build_purpose_join
	err = compiler.build(joinText)
	if err != nil {
		return nil, err
	}
	return compiler, nil
}
