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
func clearMap(data *map[string]string) {
	(*data)["X"] = "dsadsad"
}
func TestSimple(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	sql1, err := db.ModelDatasource("user").Select("username").ToSql()
	assert.NoError(t, err)
	sql2, err := db.ModelDatasource("user").Select("username f,email").ToSql()
	assert.NoError(t, err)
	assert.NotEqual(t, sql1, sql2)
}
func TestExecDataSourceUserName(t *testing.T) {

	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	dsUser := db.ModelDatasource("user")
	assert.NoError(t, err)
	dsUser.Limit(10).Select(
		`id,
		username,
		isnull(email,'') email,
		concat(username,' ',email) fullName,
		day(createdOn) day`,
	).Where("isActive=true and day>10")
	users, err := dsUser.ToDict()
	assert.NotEmpty(t, users)
	n := len(users)
	fmt.Println(n)

}
func BenchmarkExecDataSourceUserName(t *testing.B) {

	//dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	for i := 0; i < t.N; i++ {
		dsUser := db.ModelDatasource("user")
		assert.NoError(t, err)
		dsUser.Limit(10).Select(
			`id,
			username,
			isnull(email,'') email,
			concat(username,' ',email) fullName,
			day(createdOn) day`,
		).Where("isActive=true and day>10")
		users, err := dsUser.ToDict()
		assert.NoError(t, err)
		assert.NotEmpty(t, users)

	}

}

func BenchmarkDataSourceUserName(t *testing.B) {

	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	dsUser := db.ModelDatasource("user")
	assert.NoError(t, err)
	qlTest := dsUser.Limit(10).Select(
		`id,
		username,
		isnull(email,'') email,
		concat(username,' ',email) fullName,
		day(createdOn) day`,
	)
	for i := 0; i < t.N; i++ {
		qlTest.Where("isActive=true and day>10")
		dsUser.ToSql()
	}

}

/*
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDataSourceUserName$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDataSourceUserName-16    	    3742	    487157 ns/op	 1611334 B/op	     220 allocs/op
PASS
ok  	github.com/vn-go/dx/test	2.790s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDataSourceUserName$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDataSourceUserName-16    	    3866	    992752 ns/op	 4075444 B/op	     245 allocs/op
PASS
ok  	github.com/vn-go/dx/test	5.457s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDataSourceUserName$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDataSourceUserName-16    	  521468	      2503 ns/op	    4363 B/op	      20 allocs/op
PASS
ok  	github.com/vn-go/dx/test	2.692s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDataSourceUserName$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDataSourceUserName-16    	  398685	      2535 ns/op	    3962 B/op	      20 allocs/op
PASS
ok  	github.com/vn-go/dx/test	1.813s
*/
