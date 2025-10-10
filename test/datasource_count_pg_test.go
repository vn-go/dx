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
	ds := db.ModelDatasource("user")
	//count(if(id<=1,'a',id<=2,?)) ok
	ds.Select("id, concat(username,'user-''p%',12,?,'''x') nameTest", "OK") //.Where("username like 'user-''p%' and id=?", 1)
	// psql, err := ds.ToSql()
	// assert.NoError(t, err)
	// assert.NotEmpty(t, psql)
	// assert.Equal(t, len(psql.Args), 6)
	// psql, err = ds.ToSql()
	// assert.NoError(t, err)
	// assert.NotEmpty(t, psql)
	// assert.Equal(t, len(psql.Args), 6)
	// fmt.Println(psql.Sql)
	d, err := ds.ToDict()
	fmt.Println(err)
	assert.NotEmpty(t, d)

}
