package test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	dxErrors "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/test/models"
)

var sqlServerDns = "sqlserver://sa:123456@localhost?database=a001&fetchSize=10000&encrypt=disable"
var sqlDb *dx.DB

func TestFirstMssql(t *testing.T) {
	db, err := dx.Open("sqlserver", sqlServerDns)
	assert.NoError(t, err)
	defer db.Close()
	user := &models.User{}
	err = db.First(user)
	assert.NoError(t, err)
}
func BenchmarkFirstMssql(t *testing.B) {
	db, err := dx.Open("sqlserver", sqlServerDns)
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
func TestFirstFilterMssql(t *testing.T) {
	db, err := dx.Open("sqlserver", sqlServerDns)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	err = db.First(user, "username=?", "admin")
	if !errors.Is(err, &dxErrors.NotFoundErr{}) {
		assert.NoError(t, err)
	}

}
func BenchmarkFirstFilterrMssql(t *testing.B) {
	db, err := dx.Open("sqlserver", sqlServerDns)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	for i := 0; i < t.N; i++ {
		db.First(user, "username=?", "admin")
	}

}
func TestFirstbyWhereMssql(t *testing.T) {
	db, err := dx.Open("sqlserver", sqlServerDns)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := &models.User{}
	err = db.Where("username=?", "admin").First(user)
	if !errors.Is(err, &dxErrors.NotFoundErr{}) {
		assert.NoError(t, err)
	}

}
func BenchmarkFirstbyWhereMssql(t *testing.B) {
	db, err := dx.Open("sqlserver", sqlServerDns)
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
