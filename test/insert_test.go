package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestInsertUser(t *testing.T) {
	user, err := dx.NewDTO[models.User]()
	user.Username = "admin"
	assert.NoError(t, err)
	db, err := dx.Open("mysql", mySqlDsn)
	//dx.SetManagerDb("mysql", "a001")
	assert.NoError(t, err)
	err = db.Insert(user)
	assert.NoError(t, err)
}
