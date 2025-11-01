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
	query := `user(count(id) UserCount),
          role(id, name),
          from(left(user.roleId=role.id)),
          //group(role.id, role.name),
          where(UserCount > ?),  // ← Chỉ filter theo số user
          sort(UserCount desc),
          take(?)`

	err = db.DslQuery(&userInfos, query, 100, 1)
	// err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id)")
	if err != nil {
		panic(err)
	}
	t.Log(userInfos)

}
