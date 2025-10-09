package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestDeleteMySql(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user, _ := dx.NewDTO[models.User]()
	user.Username = "admin"
	db.Insert(user)
	ret := db.Delete(user, "username=?", "admin")
	assert.NoError(t, ret.Error)
}
func TestDeletePg(t *testing.T) {
	db, err := dx.Open("postgres", pgDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user, _ := dx.NewDTO[models.User]()
	user.Username = "admin"
	db.Insert(user)
	ret := db.Delete(user, "username=?", "admin")
	assert.NoError(t, ret.Error)
}
func BenchmarkDeleteMySql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	for i := 0; i < t.N; i++ {
		user, _ := dx.NewDTO[models.User]()
		user.Username = "admin"
		db.Insert(user)
		ret := db.Delete(user, "username=?", "admin")
		assert.NoError(t, ret.Error)
	}

}
