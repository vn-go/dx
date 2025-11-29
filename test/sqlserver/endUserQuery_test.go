package sqlserver

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestEndUserQuery(t *testing.T) {
	for i := 0; i < 5; i++ {
		dx.Options.ShowSql = true
		db, err := dx.Open("sqlserver", cnn)
		if err != nil {
			t.Error(err)
		}
		defer db.Close()
		qr := dx.NewEndUserQuery()
		qr = qr.Fields("user(concat(username, '''123'+?) Username),id", 1)
		// qr = qr.From("user")
		qr = qr.Filter("id = ?", 1)

		fe := qr.ToFrontEnd(db)
		fe = fe.Select("id, len(username+'''1234') Len")
		fe = fe.Filter("Len > ? and (count(id)>? or sum(len(username))=0)", 5, 1)
		query, err := fe.ToSql()
		if err != nil {
			panic(err)
		}
		fmt.Println(query.String())
		fmt.Println(query)
	}

}
func TestEndUserQueryToArray(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	qr := dx.NewEndUserQuery()
	qr = qr.Fields("user(concat(username, '''123'+?) Username),id", 1)
	// qr = qr.From("user")
	qr = qr.Filter("id = ?", 1)

	fe := qr.ToFrontEnd(db)
	fe = fe.Select("id, len(username+'''1234') Len")
	fe = fe.Filter("Len > ? and (count(id)>? or sum(len(username))=0)", 5, 1)
	//items, err := fe.ScopeQuery(t.Context())
	if err != nil {
		panic(err)
	}
	//t.Log(items)
}
func BenchmarkEndUserQuery(b *testing.B) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	expectedSql := "SELECT [T1].[id] [Id], len(concat([T1].[username], '''123' + 1) + '''1234') [Len] FROM [sys_users] [T1] WHERE [T1].[id] = 5 AND ([T1].[id] = 1) AND (len(concat([T1].[username], '''123' + 1) + '''1234') > 5) GROUP BY [T1].[id],len(concat([T1].[username], '''123' + 1) + '''1234') HAVING (count([T1].[id]) > 1 OR sum(len(concat([T1].[username], '''123' + 1))) = 0)"
	b.Run("EndUserQuery", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			qr := dx.NewEndUserQuery()
			qr = qr.Fields("user(concat(username, '''123'+?) Username),id", 1)
			// qr = qr.From("user")
			qr = qr.Filter("id = ?", 1)

			fe := qr.ToFrontEnd(db)
			fe = fe.Select("id, len(username+'''1234') Len")
			fe = fe.Filter("Len > ? and (count(id)>? or sum(len(username))=0)", 5, 1)
			query, err := fe.ToSql()
			if err != nil {
				panic(err)
			}
			assert.Equal(b, expectedSql, query.String())
			items, err := fe.ToDyanmicArrayWithContext(b.Context())
			if err != nil {
				panic(err)
			}
			b.Log(items)
		}

	})
	b.Run("EndUserQuery parallel", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				qr := dx.NewEndUserQuery()
				qr = qr.Fields("user(concat(username, '''123'+?) Username),id", 1)
				// qr = qr.From("user")
				qr = qr.Filter("id = ?", 1)

				fe := qr.ToFrontEnd(db)
				fe = fe.Select("id, len(username+'''1234') Len")
				fe = fe.Filter("Len > ? and (count(id)>? or sum(len(username))=0)", 5, 1)
				query, err := fe.ToSql()
				if err != nil {
					panic(err)
				}
				assert.Equal(b, expectedSql, query.String())
				items, err := fe.ToDyanmicArrayWithContext(b.Context())
				if err != nil {
					panic(err)
				}
				b.Log(items)
			}
		})
	})
}

/*
	Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkEndUserQuery$ github.com/vn-go/dx/test/sqlserver

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test/sqlserver
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkEndUserQuery/EndUserQuery-16         	   19462	     62982 ns/op	   95420 B/op	     298 allocs/op
BenchmarkEndUserQuery/EndUserQuery_parallel-16         	   14520	     82785 ns/op	   96269 B/op	     302 allocs/op
PASS
ok  	github.com/vn-go/dx/test/sqlserver	4.980s
------------- after optimize -----------------
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkEndUserQuery$ github.com/vn-go/dx/test/sqlserver

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test/sqlserver
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkEndUserQuery/EndUserQuery-16         	   27591	     44898 ns/op	   73912 B/op	     237 allocs/op
BenchmarkEndUserQuery/EndUserQuery_parallel-16         	   20539	     61556 ns/op	   74583 B/op	     240 allocs/op
PASS
ok  	github.com/vn-go/dx/test/sqlserver	5.424s
--- after third optimize ---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkEndUserQuery$ github.com/vn-go/dx/test/sqlserver

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test/sqlserver
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkEndUserQuery/EndUserQuery-16         	  128617	      8291 ns/op	    7231 B/op	      67 allocs/op
BenchmarkEndUserQuery/EndUserQuery_parallel-16         	  453534	      2731 ns/op	    7215 B/op	      67 allocs/op
PASS
ok  	github.com/vn-go/dx/test/sqlserver	4.691s
----- optimze 4 ----
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkEndUserQuery$ github.com/vn-go/dx/test/sqlserver

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test/sqlserver
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkEndUserQuery/EndUserQuery-16         	  169018	      8431 ns/op	    7326 B/op	      67 allocs/op
BenchmarkEndUserQuery/EndUserQuery_parallel-16         	  392367	      2895 ns/op	    7314 B/op	      67 allocs/op
PASS
ok  	github.com/vn-go/dx/test/sqlserver	4.034s
---- optimize 5 ----
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkEndUserQuery$ github.com/vn-go/dx/test/sqlserver

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test/sqlserver
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkEndUserQuery/EndUserQuery-16         	  150500	      7840 ns/op	    7151 B/op	      66 allocs/op
BenchmarkEndUserQuery/EndUserQuery_parallel-16         	  393561	      3083 ns/op	    7145 B/op	      66 allocs/op
PASS
ok  	github.com/vn-go/dx/test/sqlserver	5.829s
--- final with database exec and get rows ---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkEndUserQuery$ github.com/vn-go/dx/test/sqlserver

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test/sqlserver
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkEndUserQuery/EndUserQuery-16         	    4116	    333812 ns/op	   25733 B/op	     279 allocs/op
--- BENCH: BenchmarkEndUserQuery/EndUserQuery-16
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
    endUserQuery_test.go:84: []
	... [output truncated]
BenchmarkEndUserQuery/EndUserQuery_parallel-16         	    8193	    122724 ns/op	   34238 B/op	     308 allocs/op
--- BENCH: BenchmarkEndUserQuery/EndUserQuery_parallel-16
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
    endUserQuery_test.go:108: []
	... [output truncated]
PASS
ok  	github.com/vn-go/dx/test/sqlserver	4.583s

*/
