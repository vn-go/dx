package shorttest

import (
	"fmt"
	"testing"

	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
)

var cnn = "sqlserver://sa:123456@localhost:1433?database=hrm"

func TestQueryType(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	Query.ToSql(nil, db, "select * from user")
}
func TestQueryType1(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	Query.ToSql(nil, db, "select id from user u")
}
func TestQueryType2(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql(nil, db, "select if(user.id>10,1,0) Test from user left join department on user.roleId = department.id")
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType3(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql(nil, db, "select sum(if(user.id>10,1,0)) sum from user left join department on user.roleId = department.id")
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType4(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql(nil, db, "select sum(if(user.id>10,1,?)) sum from user left join department on user.roleId = department.id", 100)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType5(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql(nil, db, "select sum(if(user.id>10,1,?)) sum from user left join department on user.roleId+1 = department.id+?", 100, 12)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType6(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql(nil, db, "select sum(if(user.id>10,1,?)) sum from user left join department on user.roleId+1-? = department.id+? and department.name='admin'", 2, 3, 4)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType7(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql2(nil, db, `(src(user u, department d,u.roleId = d.id),u.username,d.name,u.username='admin',d.name='admin')+
									   (src(user u, department d,u.roleId = d.id),u.username,d.name,u.username='admin')`, 2, 3, 4)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType8(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql2(nil, db, `join(user u, department d,department d2,all(u.roleId) = d.id, d.id = d2.parentId),
								   	   select(u(username,emal),d(name DeptName)),
									   where(u.username='admin' and d.name='admin')`)
	//sql, err := Query.ToSql2(nil, db, `user(u,department d,u.id = d.id)`)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType9(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql2(nil, db, `join(
										query	(
													join(user u, department d,department d2,u.roleId = d.id, d.id = d2.parentId),
													select(u.username),
													where(u.code='admin')	
												) 	qr, 
										department d,department d2,u.roleId = d.id, d.id = d2.parentId),
								   	   	select(u.username),
									   	where(u.username='admin' and d.name='admin')`)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType10(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql2(nil, db, `
										from(user),
								   	   	select(username),
									   	where(username='admin' and email='admin')`)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType11(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql2(nil, db, `
										from(user),
								   	   	select(username),
									   	where(username='admin' and id=?)`, 100)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
func TestQueryType12(t *testing.T) {
	db, err := dx.Open("sqlserver", cnn)
	if err != nil {
		t.Error(err)
	}
	defer db.Close()
	sql, err := Query.ToSql2(&ColumnsScope{
		Cols: []string{"user.id"},
	}, db, `
										from(user)
								   	   	select(id+1 Id)
									   	where(username='admin' and id=?)
										Sort(id Desc)`, 100)
	if err != nil {
		panic(err)
	}
	fmt.Println(sql.Sql)
}
