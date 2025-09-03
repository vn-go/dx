package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
)

func TestGetFirstItemMssql(t *testing.T) {
	db, err := dx.Open("sqlserver", sqlServerDns)

	assert.NoError(t, err)
	user := &User{}
	err = db.First(user, "id=?", 1)
	assert.NoError(t, err)
}
