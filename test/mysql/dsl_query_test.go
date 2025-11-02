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
	dx.Options.ShowSql = true
	db, err := dx.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	userInfos := []struct {
		Id       uint64 `db:"pk;auto" json:"id"`
		Username string `db:"size:50;uk" json:"username"`
	}{}
	query := `
    subsets(incrementDetail(ItemID  ItemID, Amount  Amount)) AS inc,
    subsets(decrementDetail(ItemID  ItemID, Amount * -1  Amount))  dec,
    subsets(inc + dec)  inventory1,
    subsets(inventory1(ItemID ItemID, sum(Amount) AS Total))  inventory,
    inventory(Total, ItemID)
`
	query = `
		subsets(incrementDetail(ItemID  ItemID, Amount  Amount)) AS inc,
		subsets(decrementDetail(ItemID  ItemID, Amount * -1  Amount))  dec,
		inc+dec
		
	`
	query = `
    subsets(
        union(
            incrementDetail(ItemID, Amount),
            decrementDetail(ItemID, Amount * -1)
        )
    ) AS inventory_raw,
    inventory_raw(ItemID, Amount)
`
	query = `
		
			incrementDetail(ItemID, Amount) + decrementDetail(ItemID, Amount * -1 Amount),
			ItemID, sum(Amount) AS Total
		
	`
	// query = `

	// 		incrementDetail.Amount + decrementDetail.Amount * -1 A,
	// 		from(full(incrementDetail.itemID = decrementDetail.itemID))

	// `
	err = db.DslQuery(&userInfos, query)
	if err != nil {
		panic(err)
	}
}
