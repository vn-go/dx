package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestSelectSum(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	for i := 0; i < 5; i++ {
		ds := db.ModelDatasource("role").Select("left(code,2) C,count(id) total,day(createdOn) d,month(createdOn) m").Where("m=9 and total>0")
		ret, err := ds.ToDict()
		assert.NoError(t, err)
		assert.NotEmpty(t, ret)
	}
	//ds := db.ModelDatasource("user").Select("count(id) Total,year(createdOn) Year,createdBy").Where("total=6 and Year=2025")

}
func BenchmarkSelectSum(t *testing.B) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	t.Run("parallel", func(t *testing.B) {
		t.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ds := db.ModelDatasource("user").Select("count(id) Total,year(createdOn) Year,createdBy").Where("total=6 and Year=2025")

				ds.ToSql()

			}
		})
	})
	t.Run("No Parallel", func(t *testing.B) {
		for i := 0; i < t.N; i++ {

			ds := db.ModelDatasource("user").Select("count(id) Total,year(createdOn) Year,createdBy").Where("total=6 and createdBy='admin'")
			//_, err := ds.ToDict()
			ds.ToSql()
			assert.NoError(t, err)
			if err != nil {
				t.Fail()
			}
			//assert.NotEmpty(t, sql)
			// fmt.Println(sql.Sql)
			// data, err := ds.ToDict()
			// assert.NoError(t, err)
			// fmt.Println(data)
		}
	})

}

/*
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkSelectSum$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkSelectSum/parallel-16         	  262911	      5038 ns/op	    8536 B/op	      85 allocs/op
BenchmarkSelectSum/No_Parallel-16      	  142318	      8897 ns/op	    8443 B/op	      85 allocs/op
PASS
ok  	github.com/vn-go/dx/test	3.696s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkSelectSum$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkSelectSum/parallel-16         	  474136	      2484 ns/op	    6316 B/op	      61 allocs/op
BenchmarkSelectSum/No_Parallel-16      	  211989	      5811 ns/op	    6311 B/op	      61 allocs/op
PASS
ok  	github.com/vn-go/dx/test	2.773s
----
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkSelectSum$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkSelectSum/parallel-16         	 1326126	       916.3 ns/op	    3558 B/op	      20 allocs/op
BenchmarkSelectSum/No_Parallel-16      	  561578	      2092 ns/op	    3557 B/op	      20 allocs/op
PASS
ok  	github.com/vn-go/dx/test	3.065s
*/
/*
	SELECT COUNT(T1.id) AS Total,T1.id
		YEAR(T1.created_on) AS year
	FROM sys_users T1
	GROUP BY YEAR(T1.created_on),T1.id
	HAVING COUNT(T1.id) + 2 = ? AND T1.id>100
*/
