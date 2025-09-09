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

type BUILD int

const (
	BUILD_SELECT BUILD = iota
	BUILD_JOIN
	BUILD_WHERE
	BUILD_GROUP
	BUILD_HAVING
	BUILD_ORDER
	BUILD_LIMIT
	BUILD_OFFSET
	BUILD_FUNC
	BUILD_UPDATE
)

type exprCompileContext struct {
	Tables []string
	/*
		The purpose of this field is track table name is already in database
	*/
	schema           *map[string]bool
	Alias            map[string]string
	joinAlias        map[string]string
	AliasToDbTable   map[string]string
	AlterTableJoin   map[string]string
	Dialect          types.Dialect
	Purpose          BUILD
	stackAliasFields internal.Stack[string]
	stackAliasTables internal.Stack[string]
	paramIndex       int
}

func (e *exprCompileContext) pluralTableName(tableName string) string {
	if e.schema != nil {
		if _, ok := (*e.schema)[tableName]; ok {
			return tableName
		} else {
			if _, ok := e.Alias[tableName]; ok {
				return tableName
			} else {
				return internal.Utils.Pluralize(tableName)
			}
		}
	} else {
		if _, ok := e.Alias[tableName]; ok {
			return tableName
		} else {
			return internal.Utils.Pluralize(tableName)
		}
	}
}

type exprCompiler struct {
	Context *exprCompileContext
	Content string
}

func (e *exprCompiler) BuildSortField(selector string) error {
	e.Context.Purpose = BUILD_ORDER
	selector = internal.Helper.QuoteExpression(selector)
	sqlTest := "select *  from tmp order by " + selector
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlSelect, ok := stm.(*sqlparser.Select); ok {

		ret, err := exprs.compile(e.Context, sqlSelect.OrderBy)
		if err != nil {
			return err
		}
		e.Content = ret

	}

	return nil
}
func (e *exprCompiler) BuildSelectField(selector string) error {
	e.Context.Purpose = BUILD_SELECT
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
					e.Context.stackAliasFields.Push(sqlExpr.As.String())
				}
				if sqlExpr.Expr != nil {
					strResult, err := exprs.compile(e.Context, sqlExpr.Expr)

					if err != nil {
						return err
					}
					selectors[i] = strResult

				}
			} else {
				panic(fmt.Errorf("unsupported select type is %T", expr))
			}
		}
		e.Content = strings.Join(selectors, ", ")
	}

	return nil
}
func (e *exprCompiler) BuildSetter(stterExpr string) error {
	stterExpr = internal.Helper.QuoteExpression(stterExpr)

	sqlTest := "update test set " + stterExpr
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlUpdate, ok := stm.(*sqlparser.Update); ok {
		strResults := []string{}
		for _, expr := range sqlUpdate.Exprs {
			strResult, err := exprs.compile(e.Context, expr)
			if err != nil {
				return err
			}
			strResults = append(strResults, strResult)

		}
		e.Content = strings.Join(strResults, ", ")
	}

	return nil
}
func (e *exprCompiler) Build(joinText string) error {
	joinText = internal.Helper.QuoteExpression(joinText)

	sqlTest := "select * from " + joinText
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlSelect, ok := stm.(*sqlparser.Select); ok {

		for _, expr := range sqlSelect.From {
			strResult, err := exprs.compile(e.Context, expr)
			if err != nil {
				return err
			}
			e.Content = strResult
		}
	}

	return nil

}
func (e *exprCompiler) BuildWhere(where string) error {
	where = internal.Helper.QuoteExpression(where)
	e.Context.Purpose = BUILD_WHERE

	sqlTest := "select * from tmp where" + where
	stm, err := sqlparser.Parse(sqlTest)
	if err != nil {
		return err
	}
	if sqlSelect, ok := stm.(*sqlparser.Select); ok {
		strResult, err := exprs.compile(e.Context, sqlSelect.Where.Expr)
		// for _, expr := range sqlSelect.From {
		// 	strResult, err := exprs.compile(e.Context, expr)
		if err != nil {
			return err
		}
		e.Content = strResult
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

		err = m.DoMigrates(db) //<-- thuc hien migrate
		if err != nil {
			init.err = err //<-- loi bo bien dich khoi dong bi hong
			return
		}
		loader := m.GetLoader()                //<-- khoi tao bo loader cua migrate
		tables, err := loader.LoadAllTable(db) // <--- lay danh sach cac bang trong database va danh sach cac ban da duoc migrate
		if err != nil {
			init.err = err
			return
		}
		dialect := factory.DialectFactory.Create(db.Info.DriverName) //<-- khoi tao dialect, neu kg co dialect se kg the bien dich cua phap dung
		schema := map[string]bool{}
		for k := range tables {
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
		Context: &exprCompileContext{
			Tables:           make([]string, 0),
			schema:           &init.val.schema,
			Alias:            make(map[string]string),
			AliasToDbTable:   make(map[string]string),
			Dialect:          init.val.dialect,
			Purpose:          BUILD_SELECT,
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
	compiler.Context.Purpose = BUILD_JOIN
	err = compiler.Build(joinText)

	if err != nil {
		return nil, err
	}
	return compiler, nil
}
func ExtractTableFromJoin(joinText string) ([]string, error) {
	return internal.OnceCall(joinText, func() ([]string, error) {
		joinText = internal.Helper.QuoteExpression(joinText)

		sqlTest := "select * from tmp " + joinText
		stm, err := sqlparser.Parse(sqlTest)
		if err != nil {
			return nil, err
		}
		selectStm := stm.(*sqlparser.Select)
		m := make(map[string]bool, 0)
		ret := tabelExtractor.getTables(selectStm.From, m)
		return ret[1:], nil
	})

}

type tabelExtractorTypes struct {
}

var tabelExtractor = &tabelExtractorTypes{}

func (t *tabelExtractorTypes) getTables(node sqlparser.SQLNode, visited map[string]bool) []string {
	//sqlparser.TableExprs
	ret := []string{}
	if tableExprs, ok := node.(sqlparser.TableExprs); ok {
		for _, n := range tableExprs {
			nextTbl := t.getTables(n, visited)
			if len(nextTbl) > 0 {
				ret = append(ret, nextTbl...)
			}
		}
		return ret
	}
	if joinTableExpr, ok := node.(*sqlparser.JoinTableExpr); ok {

		nextTbl := t.getTables(joinTableExpr.LeftExpr, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		nextTbl = t.getTables(joinTableExpr.RightExpr, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		nextTbl = t.getTables(joinTableExpr.Condition, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		return ret
	}
	if aliasedTableExpr, ok := node.(*sqlparser.AliasedTableExpr); ok {
		nextTbl := t.getTables(aliasedTableExpr.As, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		nextTbl = t.getTables(aliasedTableExpr.Expr, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		return ret
	}
	//sqlparser.TableIdent
	if tableIdent, ok := node.(sqlparser.TableIdent); ok {
		if tableIdent.IsEmpty() {
			return nil
		} else {
			if _, ok := visited[tableIdent.String()]; !ok {
				visited[tableIdent.String()] = true
				ret = append(ret, tableIdent.String())
			}

		}
		return ret
	}
	//sqlparser.TableName
	if tableName, ok := node.(sqlparser.TableName); ok {
		if tableName.IsEmpty() {
			return nil
		} else {
			if _, ok := visited[tableName.Name.String()]; !ok {
				visited[tableName.Name.String()] = true
				ret = append(ret, tableName.Name.String())
			}
			return ret
		}
	}
	//sqlparser.JoinCondition
	if joinCondition, ok := node.(sqlparser.JoinCondition); ok {
		nextTbl := t.getTables(joinCondition.On, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		return ret
	}
	//*sqlparser.ComparisonExpr
	if comparisonExpr, ok := node.(*sqlparser.ComparisonExpr); ok {

		nextTbl := t.getTables(comparisonExpr.Left, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		nextTbl = t.getTables(comparisonExpr.Right, visited)
		if len(nextTbl) > 0 {
			ret = append(ret, nextTbl...)
		}
		return ret
	}
	//*sqlparser.ColName
	if colName, ok := node.(*sqlparser.ColName); ok {
		if colName.Qualifier.IsEmpty() {
			return nil
		} else {
			if _, ok := visited[colName.Qualifier.Name.String()]; !ok {
				visited[colName.Qualifier.Name.String()] = true
				ret = append(ret, colName.Qualifier.Name.String())
			}
			return ret
		}
	}

	//sqlparser.Expr
	panic(fmt.Sprintf("not implement, tabelExtractorTypes.getTables %s", `expr\expr.join.go`))

}
