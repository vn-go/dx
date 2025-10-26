package testarrayfields

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"
var mssqlDns = "sqlserver://sa:123456@localhost:1433?database=hrm"

func Test0001(t *testing.T) {
	db, err := dx.Open("sqlserver", mssqlDns)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	dep, err := dx.NewDTO[Department]()
	if err != nil {
		panic(err)
	}
	dep.Code = "A001"
	dep.Name = "XXXX"
	dep.RecordID = uuid.NewString()
	// dep.ChildrenID =
	err = db.Insert(dep)
	if err != nil {
		panic(err)
	}
}
func TestSmartyWithArrayFilter(t *testing.T) {
	//JSON_OVERLAPS(children_id, JSON_ARRAY(2,4))
	// db, err := dx.Open("mysql", dsn)
	// if err != nil {
	// 	panic(err)
	// }
	db, err := dx.Open("sqlserver", mssqlDns)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	expectedArgs := `[
   "%",
   "%"
 ]`
	expectedAccessScope := `{
  "department.childrenid": {
    "EntityName": "Department",
    "EntityFieldName": "ChildrenID"
  },
  "department.code": {
    "EntityName": "Department",
    "EntityFieldName": "Code"
  },
  "department.name": {
    "EntityName": "Department",
    "EntityFieldName": "Name"
  }
}`
	expectedSql := "SELECT [dep].[code] [Code], [dep].[name] [Name] FROM [departments] [child] join  [departments] [dep] ON [child].[children_id] like concat(@p1, [dep].[children_id], @p2)"
	sql, err := db.Smart(`
		dep(code,name,count(child.id) TotalChildren),
		sum(dep.id),
		from(
			department child,
			department dep,
			child.ChildrenId like concat('%',dep.ChildrenId,'%')  /* dep.ChildrenId is '.1.' and child.ChildrenId is '.1.2.' */
		),
		
	`, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)
	assert.Equal(t, expectedSql, sql.Query)
	assert.Equal(t, expectedAccessScope, sql.ScopeAccess.String())
	assert.Equal(t, expectedArgs, sql.Args.String())

}

func BenchmarkSmartyWithArrayFilter(t *testing.B) {

	db, err := dx.Open("sqlserver", mssqlDns)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	expectedArgs := `[
   "%",
   "%"
 ]`
	expectedAccessScope := `{
  "department.childrenid": {
    "EntityName": "Department",
    "EntityFieldName": "ChildrenID"
  },
  "department.code": {
    "EntityName": "Department",
    "EntityFieldName": "Code"
  },
  "department.name": {
    "EntityName": "Department",
    "EntityFieldName": "Name"
  }
}`
	expectedSql := "SELECT [dep].[code] [Code], [dep].[name] [Name] FROM [departments] [child] join  [departments] [dep] ON [child].[children_id] like concat(@p1, [dep].[children_id], @p2)"
	sql, err := db.Smart(`
		dep(code,name,count(child.id) TotalChildren),
		
		from(
			department child,
			department dep,
			child.ChildrenId like concat('%',dep.ChildrenId,'%')  /* dep.ChildrenId is '.1.' and child.ChildrenId is '.1.2.' */
		),
		
	`, 1)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectedSql, sql.Query)
	assert.Equal(t, expectedAccessScope, sql.ScopeAccess.String())
	assert.Equal(t, expectedArgs, sql.Args.String())

}

func TestSmartyRowset(t *testing.T) {
	//JSON_OVERLAPS(children_id, JSON_ARRAY(2,4))
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sql, err := db.Smart(`
		rowset(user(departmentId),where)
		dataset.department(count(id) Count),
		where(id=)
	`, []int{1, 2})
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Query)
}
