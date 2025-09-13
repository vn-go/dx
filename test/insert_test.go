package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	dxErr "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/test/models"
)

func TestInsertUser(t *testing.T) {
	user, err := dx.NewDTO[models.User]()
	user.Username = "user12345"
	assert.NoError(t, err)
	db, err := dx.Open("mysql", mySqlDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	err = db.Insert(user)
	if dxError, ok := err.(*dxErr.DbErr); ok {
		if dxError.ErrorType != dxErr.ERR_DUPLICATE {
			t.Fail()
		}
	} else {
		assert.NoError(t, err)
	}

}
func BenchmarkInsertUser(t *testing.B) {
	db, _ := dx.Open("mysql", mySqlDsn)
	defer db.Close()
	for i := 0; i < t.N; i++ {

		for j := 0; j < 10000; j++ {
			user, _ := dx.NewDTO[models.User]()
			user.Username = fmt.Sprintf("user-ok-%d", j+i*10)
			db.Insert(user)
			//assert.NoError(t, err)
		}

		//dx.SetManagerDb("mysql", "a001")

	}

}
func TestInsertUserWithContext(t *testing.T) {
	user, err := dx.NewDTO[models.User]()
	user.Username = "user12345"
	assert.NoError(t, err)
	db, err := dx.Open("mysql", mySqlDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	err = db.WithContext(t.Context()).Insert(user)
	if dxError, ok := err.(*dxErr.DbErr); ok {
		if dxError.ErrorType != dxErr.ERR_DUPLICATE {
			t.Fail()
		}
	} else {
		assert.NoError(t, err)
	}

}
