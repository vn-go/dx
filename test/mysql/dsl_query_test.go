package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vn-go/dx"
	_ "github.com/vn-go/dx/test/models"
)

func TestBasic(t *testing.T) {
	type Location struct {
		// Khóa chính (BẮT BUỘC)
		ID int `db:"pk;auto" json:"id"`

		// Các cột khác
		City string `db:"size:100" json:"city"`
		Code string `db:"size:50" json:"code"`

		// ... các trường cần thiết khác
		// BaseModel (Nếu bạn sử dụng BaseModel)
	}
	dx.AddModels(&Location{}) // add model to dx truoc khi open database
	var dsn string = "root:123456@tcp(127.0.0.1:3306)/hrm2"
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	userInfos := []struct {
		Users  uint64 `db:"pk;auto" json:"id"`
		RoleId string `db:"size:50;uk" json:"username"`
	}{}

	query := `user(id, username), 
          role(name as roleName), 
          department(name as deptName),
          location(city), //-- Thêm bảng thứ 4
          
          from(left(user.roleId = role.id), 
               left(user.departmentId = department.id), 
               left(department.locationId = location.id)), //-- Thêm điều kiện nối thứ 3

          where(user.id = ?)`
	err = db.DslQuery(&userInfos, query, 10, 100)
	if err != nil {
		panic(err)
	}

	/*
		bat loi truoc khi chuyen den database engine
		panic: Please add a name (alias) for the expression 'count(user.userid)'. [recovered, repanicked]

	*/
}
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
	query := `
    user(),
    role(),
    department(),
    from(
        left(user.roleId = role.id),
        left(user.departmentId = department.id)
    ),
    where(user.isActive = 1),
    sort(user.createdOn desc),
    take(?)
`
	err = db.DslQuery(&userInfos, query, 10)

	err = db.DslQuery(&userInfos, query, 100)
	// err = db.DslQuery(&userInfos, "user(id,username),where(id>=1),sort(id)")
	if err != nil {
		panic(err)
	}
	t.Log(userInfos)

}
func TestQuery2(t *testing.T) {
	//var fx models.DecrementDetail
	userInfos := []struct {
		Id       uint64 `db:"pk;auto" json:"id"`
		Username string `db:"size:50;uk" json:"username"`
	}{}
	type Admin struct {
		Id       uint64 `db:"pk;auto" json:"id"`
		Username string `db:"size:50;uk" json:"username"`
		Name     string `db:"size:100" json:"name"` // hoặc dùng Username nếu không có Name
		// ... các field khác tương tự User nếu cần
	}

	dx.AddModels(&Admin{})
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	expectexSql := "SELECT d.name, stats.userCount FROM department `d` INNER JOIN (SELECT u.userId as id, count(*) userCount FROM (SELECT user.id as userId, user.username FROM user UNION ALL SELECT admin.id as userId, admin.username FROM admin)  u) `stats` ON d.managerID = stats.id"
	query := `
		
			subsets(incrementDetail(ItemID ItemId, Amount) + 
						decrementDetail(ItemID ItemId, Amount * -1 Amount)) q1,
			subsets(q1(ItemID ItemId, sum(Amount) AS Total)) q2,
			from(item i,i.id=q2.itemId),
			i(),q2()	 		
		
	`
	query = `
    subsets(incrementDetail(ItemID ItemId, Amount) + 
                decrementDetail(ItemID ItemId, Amount * -1 Amount)) q1,
    subsets(q1(ItemID ItemId, sum(Amount) Total)) q2,
    from(item i, i.id=q2.ItemId),
    i(), q2.Total
`
	query = `
subsets(user(id userId, username) + admin(id userId, username)) u,
subsets(u(userId id),count(*) userCount) stats,  // ✅ dua count(*) userCount vao trong u(...)
from(department d, d.managerID=stats.id),         // ✅ Dùng alias mới
d.name, stats.userCount
`
	query = `
    subsets(user(id userId, username) + admin(id userId, username)) u,
    subsets(u(userId id), count(*) userCount) stats,
    from(department d, d.managerID=stats.id),
    d.name, stats.userCount
`
	sql, err := db.Compact(query)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, expectexSql, sql)
	err = db.DslQuery(&userInfos, query)
	if err != nil {
		panic(err)
	}
}

func BenchmarkTestQuery2(b *testing.B) {

	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	userInfos := []struct {
		Id       uint64 `db:"pk;auto" json:"id"`
		Username string `db:"size:50;uk" json:"username"`
	}{}
	// expectedSql := "SELECT `T`.`ItemID` `ItemID`, sum(`T`.`Amount`) `Total` FROM (  SELECT `T1`.`item_id` `ItemID`, `T1`.`amount` `Amount` FROM `increment_details` `T1`  union all  SELECT `T1`.`item_id` `ItemID`, `T1`.`amount` * {1} `Amount` FROM `decrement_details` `T1`) `T` GROUP BY `T`.`ItemID`"
	query := `
		
			incrementDetail(ItemID, Amount) + decrementDetail(ItemID, Amount * -1 as Amount),
			ItemID, sum(Amount) AS Total
		
	`
	b.Run("Dsl-union", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// sql, err := db.Smart(query)
			// if err != nil {
			// 	panic(err)
			// }
			// assert.Equal(b, expectedSql, strings.ReplaceAll(sql.Query, "\n", " "))
			err = db.DslQuery(&userInfos, query)
			if err != nil {
				panic(err)
			}
		}
	})
	b.Run("Dsl-union-parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// sql, err := db.Smart(query)
				// if err != nil {
				// 	panic(err)
				// }
				// assert.Equal(b, expectedSql, strings.ReplaceAll(sql.Query, "\n", " "))
				err = db.DslQuery(&userInfos, query)
				if err != nil {
					panic(err)
				}
			}
		})
	})

}
