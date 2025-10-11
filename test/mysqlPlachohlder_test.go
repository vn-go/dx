package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx/internal"
)

func TestMysqlReplacePlaceholdersSafe(t *testing.T) {
	sql := `select concat(a,{1},b) where a={2} and b={3} and d={2}`
	args := []any{" ", 2, 3}

	matrix, _ := internal.Helper.ExtractParamMatrix(sql)
	internal.Helper.ApplyMatrix(matrix, args)
}
func BenchmarkMysqlReplacePlaceholdersSafe(t *testing.B) {
	newArg := []any{}
	for i := 0; i < t.N; i++ {
		sql := `select concat(a,{1},b) where a={2} and b={3} and d={2}`
		args := []any{" ", 2, 3}

		matrix, _ := internal.Helper.ExtractParamMatrix(sql)
		newArg = internal.Helper.ApplyMatrix(matrix, args)

	}
	assert.NotEmpty(t, newArg)
}

/*
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkMysqlReplacePlaceholdersSafe$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkMysqlReplacePlaceholdersSafe-16    	  802236	      1255 ns/op	     533 B/op	      10 allocs/op
PASS
ok  	github.com/vn-go/dx/test	3.059s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkMysqlReplacePlaceholdersSafe$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkMysqlReplacePlaceholdersSafe-16    	 5087566	       221.9 ns/op	     128 B/op	       3 allocs/op
PASS
ok  	github.com/vn-go/dx/test	2.641s
--- sau khi khai thien if !inSingle && !inDouble && ch == '{' && i+2 < len(sql) && sql[i+1] >= '0' && sql[i+1] <= '9' {

Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkMysqlReplacePlaceholdersSafe$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkMysqlReplacePlaceholdersSafe-16    	 5999649	       189.9 ns/op	     128 B/op	       3 allocs/op
PASS
ok  	github.com/vn-go/dx/test	3.890s
*/
