package mysql

import (
	"testing"

	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"

func TestGetFirst(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	var user models.User
	err = db.First(&user)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
	err = db.First(&user, "userid=?", user.UserId)
	t.Log(user)
	if err != nil {
		t.Error(err)
	}

}
