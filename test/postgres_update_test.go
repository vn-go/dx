package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	dxErr "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/test/models"
)

func TestUpdatetUserPostgres(t *testing.T) {

	db, err := dx.Open("postgres", pgDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	user := &models.User{}
	err = db.First(user, "username!=?", "admin")
	assert.NoError(t, err)
	user.Phone = dx.Ptr("12345667")
	err = db.Update(user).Error
	if dxError, ok := err.(*dxErr.DbErr); ok {
		if dxError.ErrorType != dxErr.ERR_DUPLICATE {
			t.Fail()
		}
	} else {
		assert.NoError(t, err)
	}
}
func TestUpdatetUserWithWherePostgres(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("postgres", pgDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	// user := &models.User{}
	// err = db.First(user, "username!=?", "admin")
	assert.NoError(t, err)
	// user.Phone = dx.Ptr("12345667")
	// db.Model(&models.User{}).Select()
	err = db.Model(&models.User{}).Where("IsAdmin=?", false).Update(map[string]interface{}{
		"Phone": "123456",
	}).Error
	if dxError, ok := err.(*dxErr.DbErr); ok {
		if dxError.ErrorType != dxErr.ERR_DUPLICATE {
			t.Fail()
		}
	} else {
		assert.NoError(t, err)
	}
}
func BenchmarkUpdatetUserWithWherePostgres(t *testing.B) {

	db, err := dx.Open("postgres", pgDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	// user := &models.User{}
	// err = db.First(user, "username!=?", "admin")
	//assert.NoError(t, err)
	// user.Phone = dx.Ptr("12345667")
	// db.Model(&models.User{}).Select()

	for i := 0; i < t.N; i++ {
		err = db.Model(&models.User{}).Where("IsAdmin=", false).Update(map[string]interface{}{
			"Phone": "123456",
		}).Error
		if dxError, ok := err.(*dxErr.DbErr); ok {
			if dxError.ErrorType != dxErr.ERR_DUPLICATE {
				t.Fail()
			}
		}
	}

}
