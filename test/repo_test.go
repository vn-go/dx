package test

// type Repo struct {
// 	dx.DbContext[Repo]
// 	Users      models.User
// 	Attendance models.Attendance
// 	Department models.Department
// 	Contract   models.Contract
// }

// func TestRepo(t *testing.T) {
// 	db, err := dx.Open("mysql", mySqlDsn)
// 	assert.NoError(t, err)
// 	(&Repo{}).Get(db)
// }
// func BenchmarkRepo(t *testing.B) {
// 	db, err := dx.Open("mysql", mySqlDsn)
// 	assert.NoError(t, err)
// 	for i := 0; i < t.N; i++ {
// 		ctx := (&Repo{}).Get(db)
// 		ctx.First(ctx.Models.Users)

// 	}

// }

/*
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkRepo$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkRepo-16    	16610941	        74.38 ns/op	      72 B/op	       2 allocs/op
PASS
ok  	github.com/vn-go/dx/test	1.378s
*/
