package mysql

import (
	"testing"

	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
)

func TestQuery(t *testing.T) {
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	db.Select()
	userInfos := []struct {
		Id       uint64 `db:"pk;auto" json:"id"`
		Username string `db:"size:50;uk" json:"username"`
	}{}
	query := `user(id,username,count(role.id) RoleCount,count(email) EmailCount), EmailCount+RoleCount as total_count,
									from( left(user.roleId=role.id)),
									sort(EmailCount),
									take(?),
									/*group(role.name),*/
									where(((RoleCount>?) and (role.name like '%%admin%%')) or total_count>100)`
	query = `
				from(user u),
					crossTab(for(day(u.createdOn) Day ,1,5),
					select(count(if(u.id>1,1,0)) Total)
				),
				where(day(u.createdOn)=2)
				`
	err = db.DslQuery(&userInfos, query)
	// err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id)")
	if err != nil {
		panic(err)
	}
	t.Log(userInfos)

}
