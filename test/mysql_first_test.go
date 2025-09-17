package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestFirstMysql(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	err = db.First(user)
	assert.NoError(t, err)
}
func BenchmarkFirstMysql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	db.First(user)
	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		db.First(user)
	}

}
func TestFirstFilterMysql(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	err = db.First(user, "username!=?", "admin")
	assert.NoError(t, err)
}
func BenchmarkFirstFilterMysql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	for i := 0; i < t.N; i++ {
		db.First(user, "username=?", "admin")
	}

}
func TestFirstbyWhereMysql(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := &models.User{}
	err = db.Where("username=?", "admin").First(user)
	if dbErr := dx.Errors.IsDbError(err); dbErr != nil {
		if dbErr.ErrorType != dx.Errors.NOTFOUND {
			assert.NoError(t, err)
		}

	} else {
		assert.NoError(t, err)
	}

}
func BenchmarkFirstbyWhereMysql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := &models.User{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		db.Where("username=?", "admin").First(user)
	}

}
