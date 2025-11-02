package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	expectexSql := "SELECT * FROM item `i` INNER JOIN (SELECT q1.ItemID as ItemId, sum(q1.Amount) as Total FROM (SELECT incrementDetail.ItemID as ItemId, incrementDetail.Amount FROM incrementDetail UNION ALL SELECT decrementDetail.ItemID as ItemId, decrementDetail.Amount * -1 as Amount FROM decrementDetail)  q1) `q2` ON i.id = q2.itemId"
	query := `
		
			subsets(incrementDetail(ItemID ItemId, Amount) + 
						decrementDetail(ItemID ItemId, Amount * -1 Amount)) q1,
			subsets(q1(ItemID ItemId, sum(Amount) AS Total)) q2,
			from(item i,i.id=q2.itemId),
			i(),q2()	 		
		
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
