package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestDeleteUserMysql(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	err = db.Delete(&models.User{}, "username=?", "admin").Error
	assert.NoError(t, err)
}
