package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestJoin2Model(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users []models.User

	err = db.Joins("LEFT JOIN department d ON user.Id = department.Id").Limit(10).Find(&users)
	assert.NoError(t, err)

}
func BenchmarkJoin2Model(t *testing.B) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users []models.User
	for i := 0; i < t.N; i++ {
		db.Joins("LEFT JOIN department d ON user.Id = department.Id").Limit(10).Find(&users)
	}

}
func TestJoin2ModelThenGetFirst(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var user models.User

	err = db.Joins("LEFT JOIN department d ON user.Id = department.Id").Limit(10).First(&user)
	assert.NoError(t, err)

}
