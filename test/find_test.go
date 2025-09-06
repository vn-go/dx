package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestFindbyWhereMysql(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := []models.User{}
	total := uint64(0)
	err = db.Model(&models.User{}).Where("username!=?", 25).Count(&total)
	assert.NoError(t, err)
	err = db.Where("username!=?", "admin").Order("Id desc").Find(&user)

	assert.NoError(t, err)
}
func BenchmarkFindbyWhereMysql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	user := []models.User{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		db.Where("username!=?", "admin").Order("Id desc").Find(&user)
	}

}
