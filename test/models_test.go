package test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

func TestModel(t *testing.T) {
	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	t.Log(db)
	item := []models.User{}
	db.SelectAll(&item)
	qr := db.Model(models.User{})
	sql, _, _ := qr.GetSQL()
	fmt.Println(sql)
	data, _ := qr.Find()

	items := data.([]models.User)
	t.Log(items)

}
func BenchmarkModel(t *testing.B) {
	items := []models.User{}

	db, err := dx.Open("mysql", mySqlDsn)
	if err != nil {
		t.Fail()
	}
	t.Run("load all from model", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			qr := db.Model(models.User{}).Limit(1000)
			qr.Find()

			// if err != nil {
			// 	t.Fail()
			// }
			// items = anyItem.([]models.User)

		}
	})
	t.Run("load all from", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			db.Select("username").Limit(1000).Find(&items)

		}
	})
	assert.Greater(t, len(items), 0)

}

/*
Running tool: C:\Golang\bin\go.exe test -benchmem -run=^$ -bench ^BenchmarkModel$ github.com/vn-go/dx/test

goos: windows
goarch: amd64
pkg: github.com/vn-go/dx/test
cpu: 12th Gen Intel(R) Core(TM) i7-12650H
BenchmarkModel/load_all_from_model-16         	      38	  36389945 ns/op	10499611 B/op	  147007 allocs/op
BenchmarkModel/load_all_from-16               	      43	  31246651 ns/op	13770146 B/op	  220523 allocs/op
PASS
ok  	github.com/vn-go/dx/test	4.061s
*/
