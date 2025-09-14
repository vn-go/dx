package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestJoin2ModelPG(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users []models.User

	err = db.Joins("LEFT JOIN department d ON user.Id = D.Id").Limit(10).Find(&users)
	assert.NoError(t, err)

}

/*
goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkJoin2ModelPG-16    	   11349	    104119 ns/op	    5500 B/op	     127 allocs/op
PASS
ok  	github.com/vn-go/dx/test	2.278s
*/
func BenchmarkJoin2ModelPG(t *testing.B) {
	//dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users []models.User
	t.Run("LEFT JOIN department d ON user.Id = department.Id", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			db.Joins("LEFT JOIN department d ON user.Id = department.Id").Limit(10).Find(&users)
			/*
				Auto-correct the table name from 'department' to 'departments'.
				The ORM automatically determines the database table name based on the model name 'department' and 'user' (model names are case-insensitive)
			*/
		}
	})
	t.Run("LEFT JOIN department d ON user.Id = d.Id", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			db.Joins("LEFT JOIN department d ON user.Id = d.Id").Limit(10).Find(&users)
			/*
				Auto-correct the table name from 'department' to 'departments' and user to users an also use join alias.
				The ORM automatically determines the database table name based on the model name 'department' and 'user' (model names are case-insensitive)
			*/
		}
	})
	t.Run("LEFT JOIN department  ON user.Id = department.Id", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			db.Joins("LEFT JOIN department d ON user.Id = department.Id").Limit(10).Find(&users)
		}
	})
	t.Run("LEFT JOIN departments  ON users.Id = departments.id", func(t *testing.B) {
		for i := 0; i < t.N; i++ {
			db.Joins("LEFT JOIN departments  ON users.Id = departments.id").Limit(10).Find(&users)
		}
	})

}
func TestJoin2ModelThenGetFirstPG(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var user models.User

	err = db.Joins("LEFT JOIN department d ON user.Id = department.Id").Limit(10).First(&user)
	assert.NoError(t, err)

}
