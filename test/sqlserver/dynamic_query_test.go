package sqlserver

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/vn-go/dx"
)

func TestXxx(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	r, err := db.Smart("user(username+'''aaa' Test, sum(if(id>0,0,?)) Count),role(count(id) NumOfRoles),from(role r,user u,left(u.roleId=r.id))", 1000)
	if err != nil {
		panic(err)
	}
	fmt.Println(r.Query)
	qr := dx.NewDynamicQuery("user(sum(if(id=0,1,'abc')) Count),role(count(id) NumOfRoles)")
	qr.Join("role r,user u,left(u.roleId=r.id)")
	data, err := qr.ToArray(db)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
	bff, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bff))
}
