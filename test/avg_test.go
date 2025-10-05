package test

import (
	"fmt"
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
	ds := db.ModelDatasource("user").Where("day(CreatedOn)=27")
	sql, err := ds.ToSql()
	assert.NoError(t, err)
	if err != nil {
		t.Fail()
	}
	fmt.Println(sql.Sql)
	data, err := ds.ToDict()
	assert.NoError(t, err)
	fmt.Println(data)
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
				ds := db.ModelDatasource("user").Select("count(id) Total,year(createdOn) Year,createdBy").Where("total=6 and createdBy='admin'")

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
	SELECT COUNT(T1.id) AS Total,T1.id
		YEAR(T1.created_on) AS year
	FROM sys_users T1
	GROUP BY YEAR(T1.created_on),T1.id
	HAVING COUNT(T1.id) + 2 = ? AND T1.id>100
*/
