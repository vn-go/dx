package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestSelecJoinMySql(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	users := []models.User{}
	a := db.Model(&models.User{})
	b := a.Select("User.userId,user.Email, user.UserName,contract.id")
	c := b.Joins("LEFT JOIN Contract c ON user.id = c.ID and c.id=?", 12)
	e := c.Group("user.UserName", "user.Email", "User.userId", "concat(user.Email,?,user.Username)", " ")
	err = e.Find(&users)
	assert.NoError(t, err)

}
func TestSelecGroupByHaveing1TableMySql(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	data := []struct {
		Total uint64
		Email *string
	}{}
	//users := []models.User{}
	a := db.Model(&models.User{})
	b := a.Select("count(*) Total,count(id)>?,email", 4).Group("email")
	c := b.Having("count(*)>0 and email is not null")
	//e := c.Group("user.UserName", "user.Email", "User.userId", "concat(user.Email,?,user.Username)", " ")
	err = c.Find(&data)
	assert.NoError(t, err)

}
