package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestUnion(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	sql := db.Sql(`select tbl.id Id from ( select u.id ID from user u where user.username=?
					union select d.id from department d) tbl`, "admin")
	txtSql, err := sql.GetExecSql()
	assert.NoError(t, err)
	assert.NotEmpty(t, txtSql)
	x := []struct {
		Id int
	}{}
	err = sql.ScanRow(&x)
	assert.NoError(t, err)
}
func TestUnionGetSql(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	sql := db.Sql(`select u.id ID from user u where u.username='admin' union select d.id from department d`)
	txtSql, err := sql.GetExecSql()
	assert.NoError(t, err)
	assert.NotEmpty(t, txtSql)

}
func BenchmarkUnionGetSql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	// sql := db.Sql(`select tbl.id Id from ( select u.id ID from user u where user.username=?
	// 				union select d.id from department d) tbl`, "admin")
	t.ResetTimer()
	for i := 0; i < t.N; i++ {

		sql := db.Sql(`select tbl.id Id from ( select u.id ID from user u where user.username=?
		union select d.id from department d) tbl`, "admin")
		txtSql, err := sql.GetExecSql()
		assert.NoError(t, err)
		assert.NotEmpty(t, txtSql)
		x := []struct {
			Id int
		}{}
		err = sql.ScanRow(&x)
	}

}
