package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)



func TestGetIemSqlServer(t *testing.T) {
	db, err := dx.Open("sqlserver", sqlServerDns)
	assert.NoError(t, err)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	user := &models.User{}
	err = db.First(user)
	assert.NoError(t, err)
}
