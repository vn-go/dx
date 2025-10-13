package sqlserver

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/vn-go/dx"
	"github.com/vn-go/dx/test/models"
)

var cnn = "sqlserver://sa:123456@localhost:1433?database=hrm"

func TestConnectDb(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
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

}
func TestInsertUser(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	user, err := dx.NewDTO[models.User]()
	if err != nil {
		t.Error(err)
	}
	user.Username = "test"
	user.Email = dx.Ptr("test@test.com")
	user.Phone = dx.Ptr("test@test.com")
	user.IsActive = true
	user.IsSysAdmin = false
	user.CreatedBy = "admin"
	user.HashPassword = "123456"
	err = db.Insert(user)
	if err != nil {
		t.Error(err)
	}
}
func TestDataSourceFromModelMssql(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	ds := db.ModelDatasource("user")
	sql, err := ds.ToSql()
	if err != nil {
		t.Error(err)
	}
	t.Log(sql.Sql)
	t.Log(sql.Args)
	user, err := ds.ToData()
	if err != nil {
		t.Error(err)
	}
	t.Log(user)
	bff, err := jsoniter.Marshal(user)
	if err != nil {
		t.Error(err)
	}
	dictData := []map[string]any{}
	err = jsoniter.Unmarshal(bff, &dictData)
	if err != nil {
		t.Error(err)
	}
	t.Log(dictData)
}
func TestDataSourceFromModel(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	dsName := "user"
	selector := "concat(username,'-',email) ok"
	selector = "sum(if(concat(username,'-',email)='m',1,0)) Total"
	selector = "1 OK,id,sum(if(concat(username,'-',email)='m',1,0)) Total"
	filter := "count(id)>0 and email is not null"
	filter = "id>0 and email is not null and count(id)+max(id)>1000"
	filter = "id>0 and email is  null and Total>1"
	source := db.ModelDatasource(dsName)
	source.ToSql()
	if selector != "" {
		source.Select(selector)
	}
	if filter != "" {
		source.Where(filter)
	}
	data, err := source.ToDict()
	if err != nil {
		t.Error(err)
	}
	t.Log(data)
}
