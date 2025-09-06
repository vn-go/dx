package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	dxErr "github.com/vn-go/dx/errors"
	"github.com/vn-go/dx/test/models"
)

func InsertFailThenUpdateUserMysql(t *testing.T) {
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
func BenchmarkInsertFailThenUpdateUserMysql(t *testing.B) {
	user, err := dx.NewDTO[models.User]()
	user.Username = "admin"
	assert.NoError(t, err)
	db, err := dx.Open("mysql", mySqlDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)

	t.ReportAllocs()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		// db.Update(user)
		// db.Insert(user)

		err = db.Insert(user)
		if dxError, ok := err.(*dxErr.DbErr); ok {
			if dxError.ErrorType == dxErr.ERR_DUPLICATE {
				user.Username = "admin1"
				db.Update(user)

			}
		}
	}

}

func TestUpdateUserFieldMysql(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	assert.NoError(t, err)
	defer db.Close()
	user := models.User{}
	err = db.First(&user, "username=?", "admin42")
	assert.NoError(t, err)
	selector := db.Model(&user).Select("email", "phone")
	result := selector.Update(&struct {
		Email string
		Phone string
	}{
		Email: "dsada-dsada-dasdasd",
		Phone: "3213-3213-3213-312",
	})
	assert.NotEmpty(t, result)

}
func BenchmarkUpdateUserFieldMysql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	assert.NoError(t, err)
	defer db.Close()
	user := models.User{}
	err = db.First(&user, "username=?", "admin42")
	assert.NoError(t, err)
	selector := db.Model(&user).Select("email", "phone")
	dataUpdate := &struct {
		Email string
		Phone string
	}{
		Email: "dsada-dsada-dasdasd",
		Phone: "3213-3213-3213-312",
	}
	t.ResetTimer()
	t.ReportAllocs()
	for i := 0; i < t.N; i++ {
		selector.Update(dataUpdate)
	}

}
