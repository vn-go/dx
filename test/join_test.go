package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestJoin2Model(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	var users []models.User
	err = db.Joins("INNER JOIN department ON user.Id = department.Id").Find(&users)
	assert.NoError(t, err)
	fmt.Println(users)
}
