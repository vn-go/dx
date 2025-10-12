package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestDsCounntPg(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	// db, err := dx.Open("mysql", mySqlDsn)
	// if err != nil {
	// 	t.Fail()
	// }
	ds := db.ModelDatasource("user")
	//count(if(id<=1,'a',id<=2,?)) ok
	ds.Select(
		//"id, count(concat(username,'user-''p%',?,?,'''x')) nameTest, concat(username,'-',email) name2", "b", "OK",
		"if(left(username,3)='admin',?,?) name", 12.3, 14.4,
	) //.Where("username like 'user-''p%' and id=?", 1)
	psql, err := ds.ToSql()
	assert.NoError(t, err)
	assert.NotEmpty(t, psql.Sql)

	d, err := ds.ToDict()
	fmt.Println(err)
	assert.NoError(t, err)
	t.Log(d)

}
func BenchmarkDsCounntPg(b *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		b.Fail()
	}
	b.Run("single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {

			ds := db.ModelDatasource("user")
			//count(if(id<=1,'a',id<=2,?)) ok
			//ds.Select("id, concat(username,'user-''p%',?,?,'''x') nameTest", "b", "OK").Where("username like 'user-''p%' and id=?", 1)
			ds.Select(
				//"id, count(concat(username,'user-''p%',?,?,'''x')) nameTest, concat(username,'-',email) name2", "b", "OK",
				"if(left(username,3)='admin',?,?) name", 12.3, 14.4,
			) //.Where("username like 'user-''p%' and id=?", 1)
			ds.ToData()
		}
	})
	b.Run("paralell", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ds := db.ModelDatasource("user")
				//count(if(id<=1,'a',id<=2,?)) ok
				// ds.Select("id, concat(username,'user-''p%',?,?,'''x') nameTest", "b", "OK").Where("username like 'user-''p%' and id=?", 1)
				// ds.ToSql()
				ds.Select(
					//"id, count(concat(username,'user-''p%',?,?,'''x')) nameTest, concat(username,'-',email) name2", "b", "OK",
					"if(left(username,3)='admin',?,?) name", 12.3, 14.4,
				) //.Where("username like 'user-''p%' and id=?", 1)
				ds.ToData()
			}

		})

	})
}

/*
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDsCounntPg$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDsCounntPg/single-16         	    6260	    161582 ns/op	   22104 B/op	     283 allocs/op
BenchmarkDsCounntPg/paralell-16       	    4969	    284332 ns/op	   26051 B/op	     290 allocs/op
PASS
ok  	github.com/vn-go/dx/test	4.872s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDsCounntPg$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDsCounntPg/single-16         	    6951	    201156 ns/op	   22691 B/op	     314 allocs/op
BenchmarkDsCounntPg/paralell-16       	    4070	    290696 ns/op	   27177 B/op	     321 allocs/op
PASS
ok  	github.com/vn-go/dx/test	4.145s
----
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDsCounntPg$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDsCounntPg/single-16         	    6932	    168889 ns/op	   24169 B/op	     345 allocs/op
BenchmarkDsCounntPg/paralell-16       	    4132	    289348 ns/op	   27295 B/op	     351 allocs/op
PASS
ok  	github.com/vn-go/dx/test	4.874s
---
 Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDsCounntPg$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDsCounntPg/single-16         	    5900	    188610 ns/op	   22285 B/op	     260 allocs/op
BenchmarkDsCounntPg/paralell-16       	    2782	    493880 ns/op	   28443 B/op	     271 allocs/op
PASS
ok  	github.com/vn-go/dx/test	6.125s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkDsCounntPg$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkDsCounntPg/single-16         	    6302	    167549 ns/op	   12864 B/op	     264 allocs/op
BenchmarkDsCounntPg/paralell-16       	    3372	    305192 ns/op	   16154 B/op	     271 allocs/op
PASS
ok  	github.com/vn-go/dx/test	3.654s
*/
