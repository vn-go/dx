package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestCountyWhereMysql(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	total := uint64(0)
	err = db.Model(&models.User{}).Where("username!=?", 25).Count(&total)
	assert.NoError(t, err)
}
func BenchmarkCountyWhereMysql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()

	total := uint64(0)
	for i := 0; i < t.N; i++ {
		db.Model(&models.User{}).Where("username!=?", 25).Count(&total)
	}

}
