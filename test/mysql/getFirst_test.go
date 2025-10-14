package mysql

import (
	"testing"

	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"

func TestGetFirst(t *testing.T) {
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	var user models.User
	err = db.First(&user)
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
	err = db.First(&user, "userid=?", user.UserId)
	t.Log(user)
	if err != nil {
		t.Error(err)
	}

}
func TestDataSourceFromModel(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	ds := db.ModelDatasource("user").Select("count(id) as Count,sum(id) Sum")
	ds = ds.Where("Count*Sum >? and username='admin'", 100)
	sql, err := ds.ToDict()
	if err != nil {
		t.Error(err)
	}
	t.Log(sql)
	ds1 := db.ModelDatasource("xxx").Select("count(id) as Count,sum(id) Sum")
	ds1 = ds1.Where("Count*Sum >? and username='admin'", 100)
	sq2, err := ds1.ToDict()
	if err != nil {
		t.Error(err)
	}
	t.Log(sq2)
}
