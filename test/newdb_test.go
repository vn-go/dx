package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestMySqlNewDb(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fatal(err)
	}
	newDB, err := db.NewDB("test00000001")
	assert.NoError(t, err)
	assert.NotNil(t, newDB)
}
func BenchmarkMySqlNewDb(b *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		newDB, err := db.NewDB("test00000002")
		assert.NoError(b, err)
		assert.NotNil(b, newDB)
	}

}
func TestMySqlGetAndUpdate(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fatal(err)
	}
	newDB, err := db.NewDB("test00000001")
	assert.NoError(t, err)
	assert.NotNil(t, newDB)
	user, err := dx.NewThenSetDefaultValues(func() (*models.User, error) {
		return &models.User{
			Username: "admin",
		}, nil
	})
	assert.NoError(t, err)
	err = db.Insert(user)
	assert.NoError(t, err)

}
