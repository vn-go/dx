package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestSelectAllMySql(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	users := []models.User{}
	err = db.SelectAll(&users)
	assert.NoError(t, err)
	fmt.Println(len(users))
}
func BenchmarkSelectAllMySql(t *testing.B) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	defer db.Close()
	users := []models.User{}
	for i := 0; i < t.N; i++ {
		db.SelectAll(&users)

	}

}
