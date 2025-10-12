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
		"if(username='admin',0,1) name",
	) //.Where("username like 'user-''p%' and id=?", 1)
	psql, err := ds.ToSql()
	assert.NoError(t, err)
	assert.NotEmpty(t, psql.Sql)
	fmt.Println(psql.Sql)
	// assert.Equal(t, len(psql.Args), 6)
	psql, err = ds.ToSql()
	assert.NoError(t, err)
	// assert.NotEmpty(t, psql)
	// assert.Equal(t, len(psql.Args), 6)
	// fmt.Println(psql.Sql)
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
			ds.Select("id, concat(username,'user-''p%',?,?,'''x') nameTest", "b", "OK").Where("username like 'user-''p%' and id=?", 1)
			ds.ToSql()
		}
	})
	b.Run("paralell", func(b *testing.B) {
		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				ds := db.ModelDatasource("user")
				//count(if(id<=1,'a',id<=2,?)) ok
				ds.Select("id, concat(username,'user-''p%',?,?,'''x') nameTest", "b", "OK").Where("username like 'user-''p%' and id=?", 1)
				ds.ToSql()
			}

		})

	})
}
