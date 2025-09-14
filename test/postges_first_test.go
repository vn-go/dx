package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestFirstPg(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	err = db.Delete(&models.User{}, "username=?", "admin").Error
	assert.NoError(t, err)
	user, err := dx.NewThenSetDefaultValues(func() (*models.User, error) {
		return &models.User{
			Username:     "admin",
			Email:        dx.Ptr("test.test.com"),
			HashPassword: dx.Ptr("test.test.com"),
		}, nil
	})
	assert.NoError(t, err)

	err = db.Insert(user)
	assert.NoError(t, err)
	//user := &models.User{}
	err = db.First(user)
	assert.NoError(t, err)
}
func BenchmarkFirstPg(t *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
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
func TestFirstFilterPg(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	err = db.First(user, "username=?", "admin")
	assert.NoError(t, err)
}
func BenchmarkFirstFilterPg(t *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	for i := 0; i < t.N; i++ {
		db.First(user, "username=?", "admin")
	}

}
func TestFirstbyWherePg(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := &models.User{}
	err = db.Where("username=?", "admin").First(user)

	assert.NoError(t, err)
}
func BenchmarkFirstbyWherePg(t *testing.B) {
	db, err := dx.Open("postgres", pgDsn)
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
