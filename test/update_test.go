package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	dxErr "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/test/models"
)

func TestUpdatetUser(t *testing.T) {
	user, err := dx.NewDTO[models.User]()
	user.Username = "admin"
	assert.NoError(t, err)
	db, err := dx.Open("mysql", mySqlDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	err = db.Insert(user)
	if dxError, ok := err.(*dxErr.DbErr); ok {
		if dxError.ErrorType != dxErr.ERR_DUPLICATE {
			t.Fail()
		} else {
			user.Username = "admin1"
			res := db.Update(user)
			assert.NoError(t, res.Error)
		}
	} else {
		assert.NoError(t, err)
	}

}
