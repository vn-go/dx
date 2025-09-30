package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

var hrCnn string = "root:123456@tcp(127.0.0.1:3306)/hrm"

func TestExecRawSqlSelect(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	users, err := db.ExecRawSqlSelectToDict(t.Context(), "select * from user,role")
	assert.NoError(t, err)
	assert.NotEmpty(t, users)
	n := len(users)
	fmt.Println(n)
}
func BenchmarkExecRawSqlSelect(t *testing.B) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	for i := 0; i < t.N; i++ {
		users, err := db.ExecRawSqlSelectToDict(t.Context(), "select * from user,role")
		assert.NoError(t, err)
		assert.NotEmpty(t, users)
	}

}
func TestExecDataSource(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	dsUser := db.NewDataSource("select id,username from user where id<=?", 1000)
	assert.NoError(t, err)
	dsUser.Limit(10).Where("Id>? AND contains(username, ?)", 10, "admin")
	users, err := dsUser.ToDict()
	assert.NotEmpty(t, users)
	n := len(users)
	fmt.Println(n)
}
func TestExecDataSourceWithCallFuncExpr(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	dsUser := db.NewDataSource("select id,concat(username,' ',email) fullName from user")
	assert.NoError(t, err)
	dsUser.Limit(10).Where("Id>? AND contains(fullName,?)", 1, "admin")
	users, err := dsUser.ToDict()
	assert.NotEmpty(t, users)
	n := len(users)
	fmt.Println(n)
}
func BenchmarkDataSourceWithCallFuncExpr(t *testing.B) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	for i := 0; i < t.N; i++ {
		dsUser := db.NewDataSource("select id,concat(username,' ',email) fullName from user")
		// assert.NoError(t, err)
		dsUser.Limit(10).Where("Id>? AND contains(fullName,?)", 1, "admin")
	}
	// dsUser := db.NewDataSource("select id,concat(username,' ',email) fullName from user")
	// // assert.NoError(t, err)
	// dsUser.Limit(10).Where("Id>? AND contains(fullName,?)", 1, "admin")
	// users, err := dsUser.ToDict()
	// assert.NotEmpty(t, users)
	// n := len(users)
	// fmt.Println(n)
}
func TestExecDataSourceUser(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	dsUser := db.NewDataSource(models.User{})
	assert.NoError(t, err)
	dsUser.Limit(10).Where("Id>?", 1)
	users, err := dsUser.ToDict()
	assert.NotEmpty(t, users)
	n := len(users)
	fmt.Println(n)
}
func TestExecDataSourceUserName(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	dsUser := db.ModelDatasource("user")
	assert.NoError(t, err)
	dsUser.Limit(10).Where("Id>?", 1)
	users, err := dsUser.ToDict()
	assert.NotEmpty(t, users)
	n := len(users)
	fmt.Println(n)

}
