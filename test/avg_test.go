package test

import (
	//"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func Add(a *[]int) {
	*a = append(*a, 1)
}
func testBool(a *bool) {
	testBool2(a)
}
func testBool2(a *bool) {
	*a = true
}
func TestUnionSource(t *testing.T) {
	x := false
	testBool(&x)
	fmt.Println(x)
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}

	ds := db.DatasourceFromSql(`select name,year(createdOn) y from (select concat(name,' ','x''0001') name, createdOn createdOn from role where name='R''0001' or name=?
								union
								select name,createdOn createdOn from role where id=497
								union all
								select name,createdOn createdOn from role where id>300 and id<350
								union all
								select r.name,r.createdOn createdOn from role r left join User on role.id=user.id where r.id>7 and r.id<400) t`, "admin")

	sql, err := ds.ToSql()
	assert.NoError(t, err)
	fmt.Println(sql.Sql)
	ret, err := ds.ToDict()
	assert.NoError(t, err)
	assert.NotEmpty(t, ret)
}
func BenchmarkUnionSource(t *testing.B) {

	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	t.Run("test-001", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds := db.DatasourceFromSql(`select concat(name,' ','x''0001') name, createdOn createdOn from role where name='R''0001' or name=?
			union
			select name,createdOn createdOn from role where id=497
			union all
			select name,createdOn createdOn from role where id>300 and id<350
			union all
			select r.name,r.createdOn createdOn from role r left join User on role.id=user.id where r.id>7 and r.id<400`, "admin")

			ds.ToSql()
		}

	})

	// ret, err := ds.ToDict()
	// assert.NoError(t, err)
	// assert.NotEmpty(t, ret)
}

/*
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkUnionSource$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkUnionSource/test-001-16         	  105255	     11755 ns/op	   13663 B/op	      92 allocs/op
PASS
ok  	github.com/vn-go/dx/test	1.879s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkUnionSource$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkUnionSource/test-001-16         	  136614	      9837 ns/op	   12557 B/op	      79 allocs/op
PASS
ok  	github.com/vn-go/dx/test	2.462s
---
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkUnionSource$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkUnionSource/test-001-16         	  128601	      9035 ns/op	   12188 B/op	      76 allocs/op
PASS
ok  	github.com/vn-go/dx/test	1.660s
*/
func TestSelectSum(t *testing.T) {
	a := []int{}
	Add(&a)
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", hrCnn)
	if err != nil {
		t.Fail()
	}
	for i := 0; i < 5; i++ {
		ds := db.DatasourceFromSql("select createdOn createdOn,name from role where name!=?", "c", "A")
		//ds := db.ModelDatasource("role")
		sql, err := ds.ToSql()
		assert.NoError(t, err)
		fmt.Println(sql.Sql)
		ret, err := ds.ToDict()
		assert.NoError(t, err)
		assert.NotEmpty(t, ret)
		ds = ds.Select("name,day(createdOn) d,month(createdOn) m").Where("m=9")
		sql, err = ds.ToSql()
		assert.NoError(t, err)
		fmt.Println(sql.Sql)
		ret, err = ds.ToDict()
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
				//fmt.Println(sql.Sql)

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
